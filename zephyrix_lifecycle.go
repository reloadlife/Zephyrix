package zephyrix

import (
	"context"

	"go.uber.org/fx"
)

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
