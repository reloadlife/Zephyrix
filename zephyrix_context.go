package zephyrix

import "github.com/gin-gonic/gin"

type Context interface {
	JSON(code int, obj interface{})
}

type zephyrixContext struct {
	z          *zephyrix
	ginContext *gin.Context
}

func (z *zephyrix) newZephyrixContext(ginContext *gin.Context) *zephyrixContext {
	return &zephyrixContext{
		ginContext: ginContext,
		z:          z,
	}
}

func (z *zephyrixContext) JSON(code int, obj interface{}) {
	z.ginContext.JSON(code, obj)
}
