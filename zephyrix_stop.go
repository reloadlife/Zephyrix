package zephyrix

import (
	"context"

	"go.uber.org/fx"
)

func (z *zephyrix) Stop() error {
	if z.fxStarted.Load() { // if we're using fx, stop it gracefully
		stopCtx, cancel := context.WithTimeout(context.Background(), fx.DefaultTimeout)
		defer cancel()

		if err := z.fx.Stop(stopCtx); err != nil {
			Logger.Error("failed to stop application gracefully: %s", err)
			return err
		}
	}
	return nil
}
