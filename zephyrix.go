package zephyrix

import "context"

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
}

type zephyrix struct {
}

func NewApplication(config *Config) Zephyrix {
	return &zephyrix{}
}
