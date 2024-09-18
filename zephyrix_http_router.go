package zephyrix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// CorsConfig represents the CORS configuration for the server.
type CorsConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

// defaultGinLogFormatter returns a formatted log string for Gin's logger.
func defaultGinLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	latency := param.Latency
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}

	return fmt.Sprintf("[Zephyrix Server] |%s %3d %s| %13v | %15s |%s %-7s %s %#v %s",
		statusColor, param.StatusCode, resetColor,
		latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

// customGinLogger returns a custom Gin logger middleware.
func (z *zephyrix) customGinLogger() gin.HandlerFunc {
	formatter := defaultGinLogFormatter
	notlogged := z.config.Server.SkipLogPaths
	skip := make(map[string]struct{}, len(notlogged))

	for _, path := range notlogged {
		skip[path] = struct{}{}
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
			Request:      c.Request,
			Keys:         c.Keys,
			TimeStamp:    time.Now(),
			Latency:      time.Since(start),
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
			BodySize:     c.Writer.Size(),
			Path:         path,
		}

		if raw != "" {
			param.Path += "?" + raw
		}

		logMsg := formatter(param)

		switch {
		case param.StatusCode >= 500:
			Logger.Error(logMsg)
		case param.StatusCode >= 400:
			Logger.Warn(logMsg)
		case param.StatusCode >= 300:
			Logger.Info(logMsg)
		default:
			Logger.Debug(logMsg)
		}
	}
}

// setupHandler configures and returns the main HTTP handler for the Zephyrix server.
func (z *zephyrix) setupHandler(handlers *ZephyrixRouteHandlers, mw *ZephyrixMiddlewares) http.Handler {
	gin.SetMode(z.getGinMode())
	handler := z.createGinEngine()

	z.configureMiddleware(handler)
	z.configureCORS(handler)
	z.configureTrustedProxies(handler)
	z.assignHandler(handler)
	z.registerRoutes(handler, handlers, mw)

	z.setupProxies(handler)

	return handler
}

// getGinMode returns the appropriate Gin mode based on the log level.
func (z *zephyrix) getGinMode() string {
	if z.config.Log.Level == "debug" {
		return gin.DebugMode
	}
	return gin.ReleaseMode
}

// createGinEngine creates and configures a new Gin engine.
func (z *zephyrix) createGinEngine() *gin.Engine {
	handler := gin.New()
	handler.UseH2C = true
	viper.SetDefault("server.max_multipart_memory", 8<<20) // 8 MiB
	handler.MaxMultipartMemory = z.config.Server.MaxMultipartMemory
	return handler
}

// configureMiddleware sets up the middleware for the Gin engine.
func (z *zephyrix) configureMiddleware(handler *gin.Engine) {
	handler.Use(gin.CustomRecovery(z.panicRecovery))
	handler.Use(z.customGinLogger())
}

// panicRecovery handles panics in the Gin engine.
func (z *zephyrix) panicRecovery(c *gin.Context, err interface{}) {
	Logger.Error("Recovery from panic: %s", err)
	if z.config.Log.Level == "debug" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// configureCORS sets up CORS for the Gin engine.
func (z *zephyrix) configureCORS(handler *gin.Engine) {
	z.setDefaultCORSValues()
	age, _ := time.ParseDuration(z.config.Server.Cors.MaxAge)
	handler.Use(cors.New(cors.Config{
		AllowOrigins:     z.config.Server.Cors.AllowedOrigins,
		AllowMethods:     z.config.Server.Cors.AllowedMethods,
		AllowHeaders:     z.config.Server.Cors.AllowedHeaders,
		ExposeHeaders:    z.config.Server.Cors.ExposedHeaders,
		AllowCredentials: z.config.Server.Cors.AllowCredentials,
		MaxAge:           age,
	}))
}

// setDefaultCORSValues sets default values for CORS configuration.
func (z *zephyrix) setDefaultCORSValues() {
	viper.SetDefault("server.cors.enabled", false)
	viper.SetDefault("server.cors.allowed_origins", []string{"*"})
	viper.SetDefault("server.cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("server.cors.allowed_headers", []string{"Origin", "Content-Length", "Content-Type", "Authorization"})
	viper.SetDefault("server.cors.exposed_headers", []string{})
	viper.SetDefault("server.cors.allow_credentials", false)
	viper.SetDefault("server.cors.max_age", "12h")
}

// configureTrustedProxies sets up trusted proxies for the Gin engine.
func (z *zephyrix) configureTrustedProxies(handler *gin.Engine) {
	if len(z.config.Server.TrustedProxies) > 0 {
		if err := handler.SetTrustedProxies(z.config.Server.TrustedProxies); err != nil {
			Logger.Error("Failed to set trusted proxies: %s", err)
		}
	}
}


var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

type RouteConfig struct {
	Methods     []string `mapstructure:"methods"`
	Path        string   `mapstructure:"path"`
	Middlewares []any    `mapstructure:"middlewares"`
}

// registerRoutes registers the routes and middleware for the Gin engine.
func (z *zephyrix) registerRoutes(handler *gin.Engine, handlers *ZephyrixRouteHandlers, mw *ZephyrixMiddlewares) {
	z.mw = mw
	routes := z.config.Server.Routes

	Logger.Debug("Registering %d dependency injected routes", len(*handlers))

	for _, route := range *handlers {
		routeName := route.Name()
		routeConfig, configExists := routes[routeName]

		methods := route.Method()
		path := route.Path()
		middlewares := make([]any, 0, len(route.Handlers()))
		middlewares = append(middlewares, route.Handlers()...)

		if configExists {
			Logger.Debug("Applying configuration for route: %s", routeName)
			if len(routeConfig.Methods) > 0 {
				methods = routeConfig.Methods
			}
			if routeConfig.Path != "" {
				path = routeConfig.Path
			}
			if len(routeConfig.Middlewares) > 0 {
				middlewares = append(middlewares, routeConfig.Middlewares...)
			}
		}

		Logger.Debug("Route %s: %v %s", routeName, methods, path)

		// Filter out invalid HTTP methods
		validatedMethods := make([]string, 0, len(methods))
		for _, method := range methods {
			if validMethods[method] {
				validatedMethods = append(validatedMethods, method)
			} else {
				Logger.Warn("Invalid HTTP method for route %s: %s", routeName, method)
			}
		}

		if len(validatedMethods) > 0 {
			handler.Match(validatedMethods, path, z.convertMiddlewares(middlewares...)...)
		} else {
			Logger.Warn("No valid HTTP methods for route %s, skipping", routeName)
		}
	}
}
