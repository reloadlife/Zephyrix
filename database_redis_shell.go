package zephyrix

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
)

type redisCompleter struct {
	commands []string
}

func (rc *redisCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line[:pos])
	if len(lineStr) == 0 {
		return
	}

	wordPart := lineStr
	if strings.Contains(lineStr, " ") {
		parts := strings.Fields(lineStr)
		wordPart = parts[len(parts)-1]
	}

	for _, cmd := range rc.commands {
		if strings.HasPrefix(strings.ToUpper(cmd), strings.ToUpper(wordPart)) {
			newLine = append(newLine, []rune(cmd[len(wordPart):]))
		}
	}

	return newLine, len(wordPart)
}

func (z *zephyrix) redisShellRun(_ *cobra.Command, args []string) error {
	poolID, err := z.getPoolID(args)
	if err != nil {
		return err
	}

	redisConfig := z.config.Database.Pools[poolID].Redis
	client := z.createRedisClient(redisConfig)
	defer client.Close()

	if err := z.testRedisConnection(client); err != nil {
		return err
	}

	z.printWelcomeMessage(redisConfig.Address)

	completer := &redisCompleter{commands: z.getRedisCommands()}
	rl, err := z.createReadline(completer)
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	ctx := context.Background()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "exit" || line == "quit" {
			break
		}

		if line == "" {
			continue
		}

		if err := z.executeRedisCommand(ctx, client, line); err != nil {
			fmt.Printf("Error: %v\n", err)
			cmd := strings.Fields(line)[0]
			z.printCommandUsage(cmd)
		}
	}

	fmt.Println("Bye")
	return nil
}

func (z *zephyrix) getPoolID(args []string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	poolID, err := strconv.Atoi(args[0])
	if err != nil || poolID < 0 || poolID >= len(z.config.Database.Pools) {
		return 0, fmt.Errorf("invalid database pool: %s", args[0])
	}
	return poolID, nil
}

func (z *zephyrix) createRedisClient(config RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})
}

func (z *zephyrix) testRedisConnection(client *redis.Client) error {
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

func (z *zephyrix) printWelcomeMessage(address string) {
	fmt.Printf("Connected to Redis server at %s\n", address)
	fmt.Printf("Type 'HELP' for a list of available commands.\n")

	// todo: print this warning until the shell is complete
	//   for contrebutions:
	// 	 if you ever fixed this shell, remove this warning. thank you :)
	fmt.Println("THIS SHELL IS NOT COMPLETE. SOME WRITE COMMANDS MAY NOT WORK OR EVEN CRASH THE SHELL.")
	fmt.Println("ANY CONTREBUTIONS ARE WELCOME TO MAKE THIS SHELL MORE AND MORE USEFUL. :) Thank you.")

	fmt.Println("Type 'exit' or 'quit' to exit the shell.")
}

func (z *zephyrix) createReadline(completer *redisCompleter) (*readline.Instance, error) {
	return readline.NewEx(&readline.Config{
		Prompt:          "redis> ",
		HistoryFile:     "/tmp/zephyrix_redis_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
}
func (z *zephyrix) executeRedisCommand(ctx context.Context, client *redis.Client, input string) error {
	args := strings.Fields(input)
	if len(args) == 0 {
		return nil
	}

	if strings.ToUpper(args[0]) == "HELP" {
		z.handleHelpCommand(args[1:])
		return nil
	}

	cmd := strings.ToUpper(args[0])
	methodName, ok := redisCommandMap[cmd]
	if !ok {
		return fmt.Errorf("unsupported command: %s", cmd)
	}

	method := reflect.ValueOf(client).MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method not found for command: %s", cmd)
	}

	methodType := method.Type()
	expectedArgs := methodType.NumIn() - 1 // Subtract 1 for the context argument
	if len(args)-1 < expectedArgs {
		return fmt.Errorf("insufficient arguments for %s command. Expected %d, got %d", cmd, expectedArgs, len(args)-1)
	}

	redisArgs := make([]reflect.Value, expectedArgs+1)
	redisArgs[0] = reflect.ValueOf(ctx)
	for i := 1; i <= expectedArgs; i++ {
		if i < len(args) {
			redisArgs[i] = reflect.ValueOf(args[i])
		} else {
			// If there are not enough arguments provided, use zero values
			redisArgs[i] = reflect.Zero(methodType.In(i))
		}
	}

	results := method.Call(redisArgs)

	if len(results) > 1 && !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	if len(results) > 0 {
		fmt.Println(z.formatRedisResult(results[0].Interface()))
	}

	return nil
}

func (z *zephyrix) formatRedisResult(result interface{}) string {
	switch v := result.(type) {
	case []string:
		return strings.Join(v, "\n")
	case []interface{}:
		var formatted []string
		for _, item := range v {
			formatted = append(formatted, fmt.Sprintf("%v", item))
		}
		return strings.Join(formatted, "\n")
	case map[string]string:
		var formatted []string
		for key, value := range v {
			formatted = append(formatted, fmt.Sprintf("%s: %s", key, value))
		}
		return strings.Join(formatted, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (z *zephyrix) getRedisCommands() []string {
	commands := make([]string, 0, len(redisCommandMap))
	for cmd := range redisCommandMap {
		commands = append(commands, cmd)
	}
	return commands
}

func (z *zephyrix) handleHelpCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Available commands:")
		for cmd := range redisCommandMap {
			fmt.Println(cmd)
		}
		fmt.Println("\nType 'HELP <command>' for more information on a specific command.")
		return
	}

	cmd := strings.ToUpper(args[0])
	z.printCommandUsage(cmd)
	// You could add more detailed help information here if desired
}
