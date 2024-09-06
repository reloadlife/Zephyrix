package zephyrix

import "github.com/gin-gonic/gin"

type zephyrixContext struct {
	*gin.Context
	z *zephyrix
}

func (z *zephyrix) newZephyrixContext(ginContext *gin.Context) *zephyrixContext {
	return &zephyrixContext{
		Context: ginContext,
		z:       z,
	}
}

func (z *zephyrixContext) JSON(code int, obj interface{}) {
	z.Context.JSON(code, obj)
}
