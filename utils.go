package zephyrix

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func (z *zephyrix) convertToGinHandlerFunc(handlerFunc any) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerValue := reflect.ValueOf(handlerFunc)
		handlerType := handlerValue.Type()

		z.validateHandlerFunc(handlerType, handlerValue)

		paramType := handlerType.In(0)
		arg := z.createHandlerArg(paramType, c)

		handlerValue.Call([]reflect.Value{arg})
	}
}

func (z *zephyrix) validateHandlerFunc(handlerType reflect.Type, handlerValue reflect.Value) {
	if handlerType.Kind() != reflect.Func {
		panic(fmt.Sprintf("Handler must be a function, got %s", handlerType))
	}

	if handlerType.NumIn() != 1 {
		panic(fmt.Sprintf("Invalid handler function signature for %s. Expected func(*gin.Context) or func(zephyrix.Context)", runtime.FuncForPC(handlerValue.Pointer()).Name()))
	}
}

func (z *zephyrix) createHandlerArg(paramType reflect.Type, c *gin.Context) reflect.Value {
	switch {
	case paramType == reflect.TypeOf(&gin.Context{}):
		return reflect.ValueOf(c)
	case paramType.Kind() == reflect.Interface && paramType.Implements(reflect.TypeOf((*Context)(nil)).Elem()):
		return reflect.ValueOf(z.newZephyrixContext(c))
	default:
		Logger.Error("Invalid handler function parameter type. Expected *gin.Context or zephyrix.Context")
	}
	return reflect.Value{}
}

func (z *zephyrix) convertMiddlewares(middlewares ...any) []gin.HandlerFunc {
	ginMiddlewares := make([]gin.HandlerFunc, 0, len(middlewares))

	for _, middleware := range middlewares {
		ginMiddleware := z.convertMiddleware(middleware)
		if ginMiddleware != nil {
			ginMiddlewares = append(ginMiddlewares, ginMiddleware)
		}
	}

	return ginMiddlewares
}

func (z *zephyrix) convertMiddleware(middleware any) gin.HandlerFunc {
	switch m := middleware.(type) {
	case gin.HandlerFunc:
		return m
	case func(*gin.Context):
		return m
	case func(Context):
		return func(c *gin.Context) {
			m(z.newZephyrixContext(c))
		}
	case string:
		return z.handleStringMiddleware(m)
	default:
		return z.convertToGinHandlerFunc(middleware)
	}
}
func (z *zephyrix) handleStringMiddleware(middlewareName string) gin.HandlerFunc {
	if z.mw == nil {
		return nil
	}

	name, args := z.parseMiddlewareName(middlewareName)
	for _, mw := range *z.mw {
		if mw.Name() == name {
			handler := mw.Handler(args...)
			return z.convertToGinHandlerFunc(handler)
		}
	}

	return nil
}

func (z *zephyrix) parseMiddlewareName(fullName string) (string, []any) {
	parts := strings.SplitN(fullName, ":", 2)
	name := parts[0]

	var args []any
	if len(parts) > 1 {
		argStrings := strings.Split(parts[1], ",")
		args = make([]any, len(argStrings))
		for i, arg := range argStrings {
			args[i] = strings.TrimSpace(arg)
		}
	}

	return name, args
}
