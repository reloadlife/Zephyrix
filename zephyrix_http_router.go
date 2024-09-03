package zephyrix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

func defaultGinLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("[Zephyrix Server] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func (z *zephyrix) customGinLogger() gin.HandlerFunc {
	formatter := defaultGinLogFormatter
	notlogged := z.config.Server.SkipLogPaths
	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()
		if _, ok := skip[path]; ok {
			return
		}

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)
		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		param.Path = path

		Logger.Debug(formatter(param))
	}
}

func (z *zephyrix) setupHandler(handlers *ZephyrixRouteHandlers) http.Handler {
	if z.config.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	handler := gin.New()
	handler.UseH2C = true
	viper.SetDefault("server.max_multipart_memory", 8<<20) // 8 MiB
	handler.MaxMultipartMemory = z.config.Server.MaxMultipartMemory

	handler.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		Logger.Error("Recovery from panic: %s", err)
		if z.config.Log.Level == "debug" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}))
	handler.Use(z.customGinLogger()) // Log requests

	// Add CORS middleware

	// DEFAULT CORS VALUES:
	viper.SetDefault("server.cors.enabled", false)
	viper.SetDefault("server.cors.allowed_origins", []string{"*"})
	viper.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("server.cors.allowed_headers", []string{"Origin", "Content-Length", "Content-Type", "Authorization"})
	viper.SetDefault("server.cors.exposed_headers", []string{})
	viper.SetDefault("server.cors.allow_credentials", false)
	viper.SetDefault("server.cors.max_age", "12h")

	age, _ := time.ParseDuration(z.config.Server.Cors.MaxAge)
	handler.Use(cors.New(cors.Config{
		AllowOrigins:     z.config.Server.Cors.AllowedOrigins,
		AllowMethods:     z.config.Server.Cors.AllowedMethods,
		AllowHeaders:     z.config.Server.Cors.AllowedHeaders,
		ExposeHeaders:    z.config.Server.Cors.ExposedHeaders,
		AllowCredentials: z.config.Server.Cors.AllowCredentials,
		MaxAge:           age,
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
		handler.Match(route.Method(), route.Path(), convertMiddlewares(route.Handlers()...)...)
	}

	return handler
}
