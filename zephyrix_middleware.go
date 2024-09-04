package zephyrix

import "go.uber.org/fx"

func (z *zephyrix) RegisterMiddleware(middlewares ...any) {
	for _, mw := range middlewares {
		z.options = append(z.options, fx.Provide(asMiddleware(mw)))
	}
}
