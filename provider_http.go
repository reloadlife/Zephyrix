package zephyrix

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
)

type zephyrixServer struct {
	httpServer  *http.Server
	httpsServer *http.Server

	errChannel chan error
	config     *ServerConfig
	z          *zephyrix

	diInjectedHandlers *ZephyrixRouteHandlers
}

func httpProvider(config *Config, z *zephyrix) (*zephyrixServer, error) {
	conf := config.Server
	server := &zephyrixServer{
		httpServer: &http.Server{
			Addr:         conf.Address,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},

		errChannel: make(chan error, 2),
		config:     &conf,
		z:          z,
	}

	if conf.TLSEnabled {
		server.httpsServer = &http.Server{
			Addr:         conf.TLSAddress,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
	}

	return server, nil
}

func httpInvoke(lc fx.Lifecycle, config *Config, logger ZephyrixLogger, server *zephyrixServer, z *zephyrix, handlers *ZephyrixRouteHandlers) {
	conf := config.Server
	server.diInjectedHandlers = handlers
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Debug("Attempting to start Zephyrix Server (Web Server)")

			go server.spawnServer("http")

			if conf.TLSEnabled {
				go server.spawnServer("https")
			}

			go func() {
				for err := range server.errChannel {
					logger.Error("Server error", err)
					err = z.Stop()
					if err != nil {
						logger.Fatal("Failed to stop the server %s", err)
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Debug("Attempting to stop Zephyrix Server (Web Server)")

			if err := server.httpServer.Shutdown(ctx); err != nil {
				logger.Error("Failed to stop HTTP server", err)
			}

			if server.httpsServer != nil {
				if err := server.httpsServer.Shutdown(ctx); err != nil {
					logger.Error("Failed to stop HTTPS server", err)
				}
			}

			return nil
		},
	})
}

func (s *zephyrixServer) spawnServer(serverType string) {
	var err error

	switch serverType {
	case "http":
		s.httpServer.Handler = s.z.setupHandler(s.diInjectedHandlers)
		s.z.r.execute()
		err = s.httpServer.ListenAndServe()
	case "https":
		s.httpsServer.Handler = s.z.setupHandler(s.diInjectedHandlers)
		s.z.r.execute()
		err = s.httpsServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)

	default:
		err = fmt.Errorf("Server of type %s is not yet supported by Zephyrix.", serverType)
	}

	if err != nil && err != http.ErrServerClosed {
		s.errChannel <- err
	}
}
