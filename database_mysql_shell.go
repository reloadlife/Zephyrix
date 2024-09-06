package zephyrix

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/chzyer/readline"
	"github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type mysqlCompleter struct {
	db *sql.DB
}

func (mc *mysqlCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])
	words := strings.Fields(lineStr)
	lastWord := ""
	if len(words) > 0 {
		lastWord = strings.ToUpper(words[len(words)-1])
	}

	var candidates []string

	// SQL keywords
	keywords := []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "TABLE", "INDEX", "VIEW", "PROCEDURE", "FUNCTION", "TRIGGER", "JOIN", "INNER", "LEFT", "RIGHT", "FULL", "OUTER", "GROUP", "BY", "HAVING", "ORDER", "LIMIT", "OFFSET"}

	for _, keyword := range keywords {
		if strings.HasPrefix(keyword, lastWord) {
			candidates = append(candidates, keyword)
		}
	}

	// Table names
	if len(words) > 1 && (strings.ToUpper(words[len(words)-2]) == "FROM" || strings.ToUpper(words[len(words)-2]) == "JOIN") {
		rows, err := mc.db.Query("SHOW TABLES")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var table string
				if err := rows.Scan(&table); err == nil {
					if strings.HasPrefix(strings.ToUpper(table), lastWord) {
						candidates = append(candidates, table)
					}
				}
			}
		}
	}

	// Column names
	if len(words) > 2 && (strings.ToUpper(words[len(words)-2]) == "SELECT" || strings.ToUpper(words[len(words)-2]) == "WHERE") {
		tableName := words[len(words)-3]
		rows, err := mc.db.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var column, _ string
				var nullable, key, extra sql.NullString
				var dataType sql.NullString
				var defaultValue sql.NullString
				if err := rows.Scan(&column, &dataType, &nullable, &key, &defaultValue, &extra); err == nil {
					if strings.HasPrefix(strings.ToUpper(column), lastWord) {
						candidates = append(candidates, column)
					}
				}
			}
		}
	}

	for _, candidate := range candidates {
		newLine = append(newLine, []rune(candidate)[len(lastWord):])
	}

	return newLine, len(lastWord)
}

func (z *zephyrix) databaseShellRun(_ *cobra.Command, args []string) error {
	poolID := 0
	if len(args) > 0 {
		var err error
		poolID, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid pool id: %s. Must be a number", args[0])
		}
		if poolID < 0 || poolID >= len(z.config.Database.Pools) {
			return fmt.Errorf("invalid pool id: %d. Must be between 0 and %d", poolID, len(z.config.Database.Pools)-1)
		}
	}

	dbConfig := z.config.Database.Pools[poolID]

	cfg, err := mysql.ParseDSN(dbConfig.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN for pool %d: %w", poolID, err)
	}

	db, err := sql.Open("mysql", dbConfig.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database using pool %d: %w", poolID, err)
	}
	defer db.Close()

	fmt.Printf("Connected to MySQL server at %s using pool %d\n", cfg.Addr, poolID)
	fmt.Println("Type 'exit' or 'quit' to exit the shell.")

	currentDB := cfg.DBName
	if currentDB == "" {
		currentDB = "(none)"
	}

	completer := &mysqlCompleter{db: db}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("mysql(pool-%d) [%s]> ", poolID, currentDB),
		HistoryFile:     "/tmp/zephyrix_mysql_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)

		if line == "exit" || line == "quit" {
			break
		}

		if line == "" {
			continue
		}

		// Print the input with syntax highlighting
		z.printWithSyntaxHighlighting(line)

		// Check if the command is "USE <database>"
		if strings.HasPrefix(strings.ToUpper(line), "USE ") {
			newDB := strings.TrimSpace(line[4:])
			if err := db.Ping(); err != nil {
				fmt.Println("Error: Lost connection to MySQL server")
				return err
			}
			if _, err := db.Exec(line); err != nil {
				fmt.Printf("Error changing database: %v\n", err)
			} else {
				currentDB = newDB
				rl.SetPrompt(fmt.Sprintf("mysql(pool-%d) [%s]> ", poolID, currentDB))
				fmt.Printf("Database changed to %s\n", currentDB)
			}
			continue
		}

		rows, err := db.Query(line)
		if err != nil {
			fmt.Println("Error executing query:", err)
			continue
		}

		z.printQueryResult(rows)
		rows.Close()
	}

	fmt.Println("Bye")
	return nil
}

func (z *zephyrix) printWithSyntaxHighlighting(input string) {
	lexer := lexers.Get("sql")
	style := styles.Get("monokai")
	formatter := formatters.Get("terminal256")

	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		fmt.Println("Error tokenizing input:", err)
		return
	}

	err = formatter.Format(os.Stdout, style, iterator)
	if err != nil {
		fmt.Println("Error formatting input:", err)
		return
	}
	fmt.Println() // Add a newline after the highlighted input
}

func (z *zephyrix) printQueryResult(rows *sql.Rows) {
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("Error getting columns:", err)
		return
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(columns)

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			return
		}

		var rowData []string
		for _, col := range values {
			if col == nil {
				rowData = append(rowData, "NULL")
			} else {
				rowData = append(rowData, string(col))
			}
		}
		table.Append(rowData)
	}

	table.Render()

	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating rows:", err)
	}
}
