package zephyrix

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
)

func convertToGinHandlerFunc(handlerFunc any) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerValue := reflect.ValueOf(handlerFunc)
		handlerType := handlerValue.Type()

		if handlerType.Kind() != reflect.Func {
			panic(fmt.Sprintf("Handler must be a function, got %s", handlerType))
		}

		if handlerType.NumIn() != 1 {
			panic(fmt.Sprintf("Invalid handler function signature for %s. Expected func(*gin.Context) or func(zephyrix.Context)", runtime.FuncForPC(handlerValue.Pointer()).Name()))
		}

		paramType := handlerType.In(0)
		var arg reflect.Value

		switch {
		case paramType == reflect.TypeOf(&gin.Context{}):
			arg = reflect.ValueOf(c)
		case paramType.Kind() == reflect.Interface && paramType.Implements(reflect.TypeOf((*Context)(nil)).Elem()):
			zephyrixCtx := newZephyrixContext(c)
			arg = reflect.ValueOf(zephyrixCtx)
		default:
			panic(fmt.Sprintf("Invalid handler function parameter type for %s. Expected *gin.Context or zephyrix.Context", runtime.FuncForPC(handlerValue.Pointer()).Name()))
		}

		handlerValue.Call([]reflect.Value{arg})
	}
}

func convertMiddlewares(middlewares ...any) []gin.HandlerFunc {
	ginMiddlewares := make([]gin.HandlerFunc, 0, len(middlewares))
	for _, middleware := range middlewares {
		switch m := middleware.(type) {
		case gin.HandlerFunc:
			ginMiddlewares = append(ginMiddlewares, m)
		case func(*gin.Context):
			ginMiddlewares = append(ginMiddlewares, m)
		case func(Context):
			ginMiddlewares = append(ginMiddlewares, func(c *gin.Context) {
				m(newZephyrixContext(c))
			})
		default:
			ginMiddlewares = append(ginMiddlewares, convertToGinHandlerFunc(middleware))
		}
	}
	return ginMiddlewares
}
