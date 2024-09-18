package zephyrix

import (
	"context"
	"sync/atomic"

	"github.com/latolukasz/beeorm/v3"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mamad.dev/zephyrix/models"
	"go.uber.org/fx"
)

type zephyrix struct {
	cobraInstance *cobra.Command
	config        *Config
	viper         *viper.Viper

	// Zephyrix FX (uber-fx)
	// this will be using the uber-go/fx under the hood.
	fx        *fx.App
	fxStarted atomic.Bool
	options   []fx.Option

	c  context.Context
	db *beeormEngine

	r  *zephyrixRouter
	mw *ZephyrixMiddlewares

	crond *cron.Cron
}

var serverGroup = &cobra.Group{
	ID:    "server",
	Title: "Server Commands",
}

var dbGroup = &cobra.Group{
	ID:    "db",
	Title: "Database Commands",
}

func (z *zephyrix) preInit() {
	z.fx = fx.New(z.options...)
}

func NewApplication() Zephyrix {
	applicationContext, cancel := context.WithCancel(context.Background())

	cobraInstance := &cobra.Command{
		Use:   "zephyrix",
		Short: "Zephyrix is a web framework",
		Long:  "Zephyrix is a web framework, https://github.com/reloadlife/zephyrix",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				Logger.Fatal("Failed to show help: %s", err)
			}
		},
	}

	cobraInstance.PersistentFlags().StringVarP(&configFilePath, "config", "c", "zephyrix.yaml", "config file path")
	cobraInstance.AddGroup(serverGroup, dbGroup)

	z := &zephyrix{
		cobraInstance: cobraInstance,
		config:        &Config{},
		viper:         viper.New(),
		c:             applicationContext,
	}
	cobra.OnInitialize(z.initConfig)
	// any additional commands will be registered here and in the actual application later on
	// TODO implement the function that will register the commands from outside zephyrix package.

	z.options = append(z.options, fx.Provide(func() ZephyrixLogger {
		return Logger
	}))
	z.options = append(z.options, fx.Provide(func() *Config {
		return z.config
	}))
	z.options = append(z.options, fx.Provide(func() *zephyrix {
		return z
	}))

	// provide the http server but never invoke it here
	z.options = append(z.options, fx.Provide(
		serverProvide,
		fx.Annotate(
			router,
			fx.ParamTags(`group:"zephyrix_router_http_fx"`),
		),
		fx.Annotate(
			mw,
			fx.ParamTags(`group:"zephyrix_mw_http_fx"`),
		)),
	)

	z.db = beeormProvider()
	z.db.RegisterEntity(&models.AuditLogEntity{})
	z.db.RegisterEntity(&models.SessionEntity{})
	
	z.options = append(z.options, fx.Provide(func() *beeormEngine {
		return z.db
	}))
	z.options = append(z.options, fx.Provide(func(bee *beeormEngine) beeorm.Engine {
		return bee.GetEngine()
	}))
	z.options = append(z.options, fx.Provide(func(config *Config, orm *beeormEngine) *AuditLogger {
		l, err := NewAuditLogger(config, orm)
		if err != nil {
			Logger.Fatal("Failed to create audit logger: %s", err)
		}
		return l
	}))
	z.options = append(z.options, fx.Provide(NewRateLimiter))
	z.options = append(z.options, fx.Invoke(invokeRateLimiter))

	z.crond = cron.New(cron.WithSeconds())
	z.options = append(z.options, fx.Invoke(z.scheduleInvoke))
	z.options = append(z.options, AuthProviderModule())

	// HTTP SERVER COMMANDS

	serveCommand := &cobra.Command{
		GroupID: serverGroup.ID,
		Use:     "serve",
		Short:   "Start the Zephyrix server",
		Long:    "Start the Zephyrix server, web server",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			z.options = append(z.options, fx.Invoke(beeormInvoke))
		},
		RunE: z.serveRun,
	}
	cobraInstance.AddCommand(serveCommand)

	// DATABASE COMMANDS

	dbCommand := &cobra.Command{
		GroupID: dbGroup.ID,
		Use:     "database",
		Short:   "Database related commands",
		Long:    "Commands for database operations",
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			defer cancel()
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			z.options = append(z.options, fx.Invoke(beeormInvoke))
		},
	}

	migrateCommand := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Run database migrations, to match the schema with the models, this will create tables, columns, indexes, etc., and will DROP any existing tables that doesnt match the schema",
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			defer cancel()
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			z.options = append(z.options, fx.Invoke(beeormInvoke))
		},
		RunE: z.migrationRun,
	}
	migrateCommand.PersistentFlags().BoolVarP(&runUnsafeMigrations, "unsafe-migrations", "f", false, "Run unsafe migrations")

	mysqlShellCommand := &cobra.Command{
		Use:   "mysql-shell [pool_id]",
		Short: "Open a MySQL shell",
		Long:  "Open an interactive MySQL shell to execute queries",
		RunE:  z.databaseShellRun,
	}

	redisShellCommand := &cobra.Command{
		Use:   "redis-shell [pool_id]",
		Short: "Open a Redis shell",
		Long:  "Open an interactive Redis shell to execute commands",
		RunE:  z.redisShellRun,
	}

	dbCommand.AddCommand(migrateCommand)
	dbCommand.AddCommand(mysqlShellCommand)
	dbCommand.AddCommand(redisShellCommand)
	z.cobraInstance.AddCommand(dbCommand)
	return z
}
