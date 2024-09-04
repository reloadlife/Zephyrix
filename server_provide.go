package zephyrix

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
)

// zephyrixServer represents the main server structure for Zephyrix.
type zephyrixServer struct {
	servers         map[serverType]*http.Server
	errChannel      chan error
	config          *parsedServerConfig
	z               *zephyrix
	logger          ZephyrixLogger
	handlers        *ZephyrixRouteHandlers
	middlewares     *ZephyrixMiddlewares
	shutdownChan    chan struct{}
	certRenewTicker *time.Ticker
	challengeServer *challengeServer
}

// newZephyrixServer creates and initializes a new zephyrixServer instance.
func newZephyrixServer(config *parsedServerConfig, z *zephyrix, logger ZephyrixLogger) *zephyrixServer {
	return &zephyrixServer{
		servers:      make(map[serverType]*http.Server),
		errChannel:   make(chan error, 2),
		config:       config,
		z:            z,
		logger:       logger,
		shutdownChan: make(chan struct{}),
	}
}

// serverProvide is a factory function that creates and configures a new zephyrixServer.
// It's designed to be used with the fx dependency injection framework.
func serverProvide(config *Config, z *zephyrix, logger ZephyrixLogger) (*zephyrixServer, error) {
	parsedConfig, err := config.Server.parse(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	server := newZephyrixServer(parsedConfig, z, logger)

	if err := server.configureServers(config); err != nil {
		return nil, err
	}

	return server, nil
}

// configureServers sets up both HTTP and HTTPS servers based on the configuration.
func (s *zephyrixServer) configureServers(config *Config) error {
	if err := s.configureHTTPServer(); err != nil {
		return fmt.Errorf("failed to configure HTTP server: %w", err)
	}

	if config.Server.SSL.Enabled {
		if err := s.configureHTTPSServer(); err != nil {
			return fmt.Errorf("failed to configure HTTPS server: %w", err)
		}
	}

	return nil
}

// serverInvoke sets up the zephyrixServer with the provided configuration and lifecycle hooks.
// It's designed to be used with the fx dependency injection framework.
func serverInvoke(lc fx.Lifecycle, config *Config, logger ZephyrixLogger, server *zephyrixServer, z *zephyrix, handlers *ZephyrixRouteHandlers, mw *ZephyrixMiddlewares) {
	server.handlers = handlers
	server.middlewares = mw

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Debug("Starting Zephyrix Server")
			return server.start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Debug("Stopping Zephyrix Server")
			return server.stop(ctx)
		},
	})
}
