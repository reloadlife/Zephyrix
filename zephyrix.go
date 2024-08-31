package zephyrix

import (
	"context"
	"sync/atomic"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
}

// Zephyrix is the interface that users will see outside of the package
// it might not be wise, to expose an interface, but will see in test phases
type Zephyrix interface {
	// Start the Zephyrix server
	// This method is blocking
	// It will return an error if the server fails to start
	// It will return error, if the server crashes
	// It will return nil if the server is stopped, gracefully, by calling Stop, or by a signal
	Start(context.Context) error

	// Stop the Zephyrix server
	// This method is blocking
	// It will return an error if the server fails to stop
	// It will return nil if the server is stopped, and cleaned up successfully
	Stop() error

	// Cleanup will stop the server,
	// and clean up any resources
	// Flush any data that is in memory to the disk or to the database (if any)
	// This method is blocking until the server is stopped
	// It will return an error if the server fails to be stopped
	// or if something goes wrong during cleanup
	// It will return nil if the server is stopped, and cleaned up successfully.
	Cleanup() error

	// Database will return the Database interface
	Database() Database
}
