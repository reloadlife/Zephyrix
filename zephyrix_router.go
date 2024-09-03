package zephyrix

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type HTTPVerb string

const (
	GET     HTTPVerb = "GET"
	POST    HTTPVerb = "POST"
	PUT     HTTPVerb = "PUT"
	DELETE  HTTPVerb = "DELETE"
	PATCH   HTTPVerb = "PATCH"
	OPTIONS HTTPVerb = "OPTIONS"
	HEAD    HTTPVerb = "HEAD"
	CONNECT HTTPVerb = "CONNECT"
	TRACE   HTTPVerb = "TRACE"
)

type Router interface {
	Group(func(router Router), ...any)
	GET(relativePath string, handlerFunction any, middlewareFunctions ...any)
	POST(relativePath string, handlerFunction any, middlewareFunctions ...any)
	PUT(relativePath string, handlerFunction any, middlewareFunctions ...any)
	DELETE(relativePath string, handlerFunction any, middlewareFunctions ...any)
	PATCH(relativePath string, handlerFunction any, middlewareFunctions ...any)
	OPTIONS(relativePath string, handlerFunction any, middlewareFunctions ...any)
	HEAD(relativePath string, handlerFunction any, middlewareFunctions ...any)
	CONNECT(relativePath string, handlerFunction any, middlewareFunctions ...any)
	TRACE(relativePath string, handlerFunction any, middlewareFunctions ...any)
	Any(relativePath string, handlerFunction any, middlewareFunctions ...any)
	Match(httpMethods []HTTPVerb, relativePath string, handlerFunction any, middlewareFunctions ...any)
}

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

func convertToGinHandlerFunc(handlerFunc any) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerValue := reflect.ValueOf(handlerFunc)
		handlerType := handlerValue.Type()

		if handlerType.NumIn() != 1 || handlerType.In(0) != reflect.TypeOf(&gin.Context{}) {
			panic(fmt.Sprintf("Invalid handler function signature for %s. Expected func(*gin.Context)", runtime.FuncForPC(handlerValue.Pointer()).Name()))
		}

		args := []reflect.Value{reflect.ValueOf(c)}
		handlerValue.Call(args)
	}
}

func (z *zephyrixRouter) handleHTTPMethod(httpMethod HTTPVerb, relativePath string, handlerFunction any, middlewareFunctions ...any) {
	ginHandlerFunc := convertToGinHandlerFunc(handlerFunction)
	middlewares := convertMiddlewares(middlewareFunctions...)
	z.assign(func() {
		if z.gHandler != nil {
			z.gHandler.Handle(string(httpMethod), relativePath, append(middlewares, ginHandlerFunc)...)
		} else {
			z.handler.Handle(string(httpMethod), relativePath, append(middlewares, ginHandlerFunc)...)
		}
	})
}

func convertMiddlewares(middlewares ...any) []gin.HandlerFunc {
	ginMiddlewares := make([]gin.HandlerFunc, 0, len(middlewares))
	for _, middleware := range middlewares {
		if ginMiddleware, ok := middleware.(gin.HandlerFunc); ok {
			ginMiddlewares = append(ginMiddlewares, ginMiddleware)
		} else {
			ginMiddlewares = append(ginMiddlewares, convertToGinHandlerFunc(middleware))
		}
	}
	return ginMiddlewares
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
