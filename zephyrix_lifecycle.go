package zephyrix

import (
	"context"

	"go.uber.org/fx"
)

// fxStart starts the application
// this is not a blocking call and will return immediately
func (z *zephyrix) fxStart() error {
	z.preInit()

	z.fxStarted.Swap(true)

	startCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
	defer cancel()
	if err := z.fx.Start(startCtx); err != nil {
		Logger.Error("failed to start application: %s", err)
		return err
	}
	return nil
}
