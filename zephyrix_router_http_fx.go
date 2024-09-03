package zephyrix

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type ZephyrixRouteHandlers []ZephyrixRouteHandler

type ZephyrixRouteHandler interface {
	Method() []string
	Path() string
	Handlers() []gin.HandlerFunc
}

func asRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(ZephyrixRouteHandler)),
		fx.ResultTags(`group:"zephyrix_router_http_fx"`),
	)
}

func router(routes []ZephyrixRouteHandler) *ZephyrixRouteHandlers {
	return (*ZephyrixRouteHandlers)(&routes)
}
