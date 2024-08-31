package zephyrix

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (z *zephyrix) Start(ctx context.Context) error {
	err := z.cobraInstance.Execute()
	if err != nil {
		Logger.Fatal("Failed to start the server: %s", err)
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-z.c.Done():
		Logger.Debug("Application Finished Execution, shutting down gracefully...")
		return z.Stop()

	case <-ctx.Done():
		Logger.Debug("Received context cancellation, shutting down gracefully...")
		return z.Stop()

	case <-c:
		Logger.Debug("Received SIGINT/SIGTERM, shutting down gracefully...")
		return z.Stop()
	}
}
