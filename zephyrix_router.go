package zephyrix

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// something to do dependency injection with
func (z *zephyrix) RegisterRouteHandler(handlers ...any) {
	for _, h := range handlers {
		z.options = append(z.options, fx.Provide(asRoute(h)))
	}
}

type zephyrixRouter struct {
	z               *zephyrix
	handler         *gin.Engine
	gHandler        *gin.RouterGroup
	afterExecution  []func()
	childExecutions []func()
}

func (z *zephyrix) assignHandler(handler *gin.Engine) *zephyrixRouter {
	z.r.handler = handler
	return z.r
}

func (z *zephyrix) Router() Router {
	if z.r == nil {
		z.r = &zephyrixRouter{
			z: z,
		}
	}
	return z.r
}

func (z *zephyrixRouter) assign(after func()) {
	z.afterExecution = append(z.afterExecution, after)
}

func (z *zephyrixRouter) execute() {
	for _, after := range z.afterExecution {
		after()
	}
	for _, child := range z.childExecutions {
		child()
	}
}

func (z *zephyrixRouter) Group(g func(router Router), options ...any) {
	var relativePath string
	handlerFunctions := make([]gin.HandlerFunc, 0)
	for _, option := range options {
		switch o := option.(type) {
		case string:
			relativePath = o
		case gin.HandlerFunc:
			handlerFunctions = append(handlerFunctions, o)
		}
	}

	z.assign(func() {
		var group *gin.RouterGroup
		if z.gHandler != nil {
			group = z.gHandler.Group(relativePath, handlerFunctions...)
		} else {
			group = z.handler.Group(relativePath, handlerFunctions...)
		}
		childRouter := &zephyrixRouter{
			handler:  z.handler,
			gHandler: group,
			z:        z.z,
		}
		g(childRouter)
		z.childExecutions = append(z.childExecutions, childRouter.execute)
	})
}

func (z *zephyrixRouter) handleHTTPMethod(httpMethod HTTPVerb, relativePath string, handlerFunction any, middlewareFunctions ...any) {
	ginHandlerFunc := z.z.convertToGinHandlerFunc(handlerFunction)
	middlewares := z.z.convertMiddlewares(middlewareFunctions...)
	z.assign(func() {
		if z.gHandler != nil {
			z.gHandler.Handle(string(httpMethod), relativePath, append(middlewares, ginHandlerFunc)...)
		} else {
			z.handler.Handle(string(httpMethod), relativePath, append(middlewares, ginHandlerFunc)...)
		}
	})
}

func (z *zephyrixRouter) GET(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(GET, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) POST(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(POST, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) PUT(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(PUT, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) DELETE(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(DELETE, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) PATCH(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(PATCH, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) OPTIONS(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(OPTIONS, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) HEAD(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(HEAD, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) CONNECT(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(CONNECT, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) TRACE(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	z.handleHTTPMethod(TRACE, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) Any(relativePath string, handlerFunction any, middlewareFunctions ...any) {
	methods := []HTTPVerb{GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, CONNECT, TRACE}
	z.Match(methods, relativePath, handlerFunction, middlewareFunctions...)
}

func (z *zephyrixRouter) Match(httpMethods []HTTPVerb, relativePath string, handlerFunction any, middlewareFunctions ...any) {
	for _, method := range httpMethods {
		z.handleHTTPMethod(method, relativePath, handlerFunction, middlewareFunctions...)
	}
}
