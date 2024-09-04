package zephyrix

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// ProxyConfig represents the configuration for a reverse proxy.
type ProxyConfig struct {
	Name        string   `mapstructure:"name"`
	Address     string   `mapstructure:"address"`
	Path        []string `mapstructure:"path"`
	IgnorePath  []string `mapstructure:"ignore_path"`
	Headers     []string `mapstructure:"headers"`
	StripPrefix bool     `mapstructure:"strip_prefix"`
}

// CustomBadGatewayError represents a custom error for bad gateway responses.
type CustomBadGatewayError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// setupProxies initializes and sets up the reverse proxies based on the configuration.
func (z *zephyrix) setupProxies(handler *gin.Engine) {
	proxyMiddleware := z.createProxyMiddleware()
	handler.Use(proxyMiddleware)
}

// createProxyMiddleware creates a gin middleware for handling reverse proxies.
func (z *zephyrix) createProxyMiddleware() gin.HandlerFunc {
	proxies := make([]*httputil.ReverseProxy, len(z.config.Server.Proxies))
	var wg sync.WaitGroup
	wg.Add(len(z.config.Server.Proxies))

	for i, proxyConfig := range z.config.Server.Proxies {
		go func(i int, proxyConfig ProxyConfig) {
			defer wg.Done()
			proxy, err := z.createReverseProxy(proxyConfig)
			if err != nil {
				Logger.Error("Failed to create reverse proxy: %s", err)
				return
			}
			proxies[i] = proxy
		}(i, proxyConfig)
	}

	wg.Wait()

	return func(c *gin.Context) {
		// Check if the route already exists in the Gin router
		if z.routeExists(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		for i, proxyConfig := range z.config.Server.Proxies {
			if z.shouldProxy(c.Request.URL.Path, proxyConfig) {
				if proxyConfig.StripPrefix {
					c.Request.URL.Path = z.stripPrefix(c.Request.URL.Path, proxyConfig.Path)
				}
				proxies[i].ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// If no proxy matches, proceed with the next handler
		c.Next()
	}
}

// routeExists checks if a route already exists in the Gin router.
func (z *zephyrix) routeExists(method, path string) bool {
	// This is a simplified check. You may need to implement a more
	// comprehensive check based on your router's structure.
	for _, route := range z.r.handler.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

// createReverseProxy creates a new reverse proxy for the given configuration.
func (z *zephyrix) createReverseProxy(proxyConfig ProxyConfig) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(proxyConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy address: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		z.modifyRequest(req, proxyConfig)
	}
	proxy.ErrorHandler = z.customErrorHandler
	Logger.Debug("Created reverse proxy for %s at %s", proxyConfig.Name, proxyConfig.Address)
	return proxy, nil
}

// shouldProxy determines if a request should be proxied based on its path and the proxy configuration.
func (z *zephyrix) shouldProxy(path string, config ProxyConfig) bool {
	// First, check if the path is in the ignore list
	for _, ignorePath := range config.IgnorePath {
		if strings.HasPrefix(path, ignorePath) {
			return false
		}
	}

	// Then check if the path should be proxied
	for _, proxyPath := range config.Path {
		if strings.HasPrefix(path, strings.TrimSuffix(proxyPath, "/*")) {
			return true
		}
	}

	return false
}

// stripPrefix removes the proxy path prefix from the request path.
func (z *zephyrix) stripPrefix(path string, prefixes []string) string {
	for _, prefix := range prefixes {
		trimmedPrefix := strings.TrimSuffix(prefix, "/*")
		if strings.HasPrefix(path, trimmedPrefix) {
			return "/" + strings.TrimPrefix(path, trimmedPrefix)
		}
	}
	return path
}

// modifyRequest modifies the incoming request based on the proxy configuration.
func (z *zephyrix) modifyRequest(req *http.Request, proxyConfig ProxyConfig) {
	for _, header := range proxyConfig.Headers {
		if val := req.Header.Get(header); val != "" {
			req.Header.Set(header, val)
		}
	}
	// Additional custom logic for request modification can be added here
}

// customErrorHandler handles errors that occur during proxying.
func (z *zephyrix) customErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	Logger.Error("Proxy error: %s", err)
	w.WriteHeader(http.StatusBadGateway)
	customError := CustomBadGatewayError{
		Code:    http.StatusBadGateway,
		Message: "Bad Gateway: The proxy server received an invalid response from an upstream server.",
	}
	if err := json.NewEncoder(w).Encode(customError); err != nil {
		Logger.Error("Failed to encode custom error: %s", err)
	}
}
