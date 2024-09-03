package zephyrix

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (z *zephyrix) setupHandler(handlers *ZephyrixRouteHandlers) http.Handler {
	if z.config.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	handler := gin.New()
	handler.Use(gin.Recovery())

	// todo: custom logger here
	handler.Use(gin.Logger()) // Log requests

	// Add CORS middleware
	handler.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if len(z.config.Server.TrustedProxies) > 0 {
		err := handler.SetTrustedProxies(z.config.Server.TrustedProxies)
		if err != nil {
			Logger.Error("Failed to set trusted proxies: %s", err)
		}
	}

	z.assignHandler(handler)

	Logger.Debug("Dependency Injected Routes: %d", len(*handlers))

	for _, route := range *handlers {
		handler.Match(route.Method(), route.Path(), route.Handlers()...)
	}

	return handler
}
