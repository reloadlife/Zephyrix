package zephyrix

import (
	"go.uber.org/fx"
)

type ZephyrixMiddlewares []ZephyrixMiddleware

type ZephyrixMiddleware interface {
	Name() string
	Handler(...any) any
}

func asMiddleware(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(ZephyrixMiddleware)),
		fx.ResultTags(`group:"zephyrix_mw_http_fx"`),
	)
}

func mw(mw []ZephyrixMiddleware) *ZephyrixMiddlewares {
	return (*ZephyrixMiddlewares)(&mw)
}
