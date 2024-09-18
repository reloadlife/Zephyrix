package zephyrix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"go.mamad.dev/zephyrix/models"
)

type AuditLogConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	AuditPool   string   `mapstructure:"audit_pool"`
	Outputs     []string `mapstructure:"outputs"`
	StoragePath string   `mapstructure:"storage_path"`
}

type AuditLogger struct {
	config     AuditLogConfig
	orm        *beeormEngine
	ormEngine  beeorm.Engine
	fileWriter *os.File
	mu         sync.Mutex
	stdout     bool
}

func NewAuditLogger(config *Config, orm *beeormEngine) (*AuditLogger, error) {
	logger := &AuditLogger{
		config: config.AuditLog,
		orm:    orm,
	}

	if err := logger.setupOutputs(); err != nil {
		return nil, err
	}

	return logger, nil
}

func (a *AuditLogger) setupOutputs() error {
	for _, output := range a.config.Outputs {
		switch {
		case strings.HasPrefix(output, "{{STORAGE}}"):
			if err := a.setupFileOutput(output); err != nil {
				return err
			}
		case output == "DATABASE":
			if err := a.setupDatabaseOutput(); err != nil {
				return err
			}
		case output == "stdout":
			a.stdout = true
		default:
			return fmt.Errorf("unknown output type: %s", output)
		}
	}
	return nil
}

func (a *AuditLogger) setupFileOutput(output string) error {
	path := strings.Replace(output, "{{STORAGE}}", a.config.StoragePath, 1)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for audit log: %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	a.fileWriter = file
	return nil
}

func (a *AuditLogger) setupDatabaseOutput() error {
	if !a.orm.HasPool(a.config.AuditPool) {
		return fmt.Errorf("database pool `%s` not found", a.config.AuditPool)
	}
	a.ormEngine = a.orm.GetEngine()
	return nil
}

func (a *AuditLogger) Log(ctx context.Context, action, userID, details string) error {
	if !a.config.Enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	timestamp := time.Now().UTC()
	logEntry := fmt.Sprintf("[%s] Action: %s, User: %s, Details: %s\n",
		timestamp.Format(time.RFC3339), action, userID, details)

	for _, output := range a.config.Outputs {
		switch {
		case output == "stdout":
			fmt.Fprint(os.Stdout, logEntry)

		case strings.HasPrefix(output, "{{STORAGE}}"):
			if _, err := a.fileWriter.WriteString(logEntry); err != nil {
				return fmt.Errorf("failed to write to audit log file: %w", err)
			}
		case output == "DATABASE":
			if err := a.logToDatabase(ctx, action, userID, details, &timestamp); err != nil {
				return fmt.Errorf("failed to log to database: %w", err)
			}
		}
	}

	return nil
}

func (a *AuditLogger) logToDatabase(ctx context.Context, action, userID, details string, timestamp *time.Time) error {
	orm := a.ormEngine.NewORM(ctx)
	al := beeorm.NewEntity[models.AuditLogEntity](orm)
	al.UserID = userID
	al.Action = action
	al.Details = details
	if timestamp != nil {
		al.CreatedAt = timestamp.UTC()
	}
	return orm.FlushAsync() // we dont want a blocking call here
}

func (a *AuditLogger) Close() error {
	if a.fileWriter != nil {
		if err := a.fileWriter.Close(); err != nil {
			return fmt.Errorf("failed to close audit log file: %w", err)
		}
	}
	return nil
}

// specific Log types (internal)

func (a *AuditLogger) logSuccessfulLogin(ctx context.Context, username string, msg ...interface{}) {
	a.Log(ctx, "login_success", username, fmt.Sprint(msg...))
}

func (a *AuditLogger) logFailedLogin(ctx context.Context, username, reason string, msg ...interface{}) {
	a.Log(ctx, "login_failure", username, fmt.Sprintf("reason: %s\n%s", reason, fmt.Sprint(msg...)))
}

func (a *AuditLogger) logAttemptLogin(ctx context.Context, username string, msg ...interface{}) {
	a.Log(ctx, "login_attempt", username, fmt.Sprint(msg...))
}

func (a *AuditLogger) logSuccessfulMFA(ctx context.Context, username, method string, msg ...interface{}) {
	a.Log(ctx, "mfa_success", username, fmt.Sprintf("method: %s\n%s", method, fmt.Sprint(msg...)))
}

func (a *AuditLogger) logFailedMFA(ctx context.Context, username, method string, msg ...interface{}) {
	a.Log(ctx, "mfa_failure", username, fmt.Sprintf("method: %s\n%s", method, fmt.Sprint(msg...)))
}
