package zephyrix

import "github.com/gin-gonic/gin"

type Context interface {
	JSON(code int, obj interface{})
}

type zephyrixContext struct {
	ginContext *gin.Context
}

func newZephyrixContext(ginContext *gin.Context) *zephyrixContext {
	return &zephyrixContext{
		ginContext: ginContext,
	}
}

func (z *zephyrixContext) JSON(code int, obj interface{}) {
	z.ginContext.JSON(code, obj)
}
