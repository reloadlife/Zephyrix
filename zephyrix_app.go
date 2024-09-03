package zephyrix

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var serverGroup = &cobra.Group{
	ID:    "server",
	Title: "Server Commands",
}

var dbGroup = &cobra.Group{
	ID:    "db",
	Title: "Database Commands",
}

func (z *zephyrix) preInit() {
	// initialize the FXapplication
	// this will be utilized to manage
	// the application lifecycle
	// and dependency injection
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
		httpProvider,
		fx.Annotate(
			router,
			fx.ParamTags(`group:"zephyrix_router_http_fx"`),
		)),
	)

	z.db = beeormProvider()
	z.options = append(z.options, fx.Provide(func() *beeormEngine {
		return z.db
	}))
	

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
	migrateCommand := &cobra.Command{
		GroupID: dbGroup.ID,
		Use:     "migrate",
		Short:   "Run database migrations",
		Long:    "Run database migrations, to match the schema with the models, this will create tables, columns, indexes, etc., and will DROP any existing tables that doesnt match the schema",
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			defer cancel()
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			z.options = append(z.options, fx.Invoke(beeormInvoke))
		},
		RunE: z.migrationRun,
	}
	migrateCommand.PersistentFlags().BoolVarP(&runUnsafeMigrations, "unsafe-migrations", "f", false, "Run unsafe migrations")
	cobraInstance.AddCommand(migrateCommand)
	return z
}
