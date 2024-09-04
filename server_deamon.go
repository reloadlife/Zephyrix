package zephyrix

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// start initializes and starts all server components.
// It spawns HTTP and HTTPS servers, sets up error monitoring,
// and initializes certificate renewal if AutoSSL is enabled.
func (s *zephyrixServer) start(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(s.servers))

	for _st, srv := range s.servers {
		wg.Add(1)
		go func(st serverType, server *http.Server) {
			defer wg.Done()
			if err := s.spawnServer(st, server); err != nil {
				errChan <- err
			}
		}(_st, srv)
	}

	go s.monitorErrors(ctx)

	if s.config.SSL.Enabled {
		if s.config.SSL.AutoSSL {
			go s.startCertificateRenewalMonitor(ctx)
		}
		if s.config.RedirectToHTTPS {
			s.SetupHTTPRedirect()
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server startup error: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// stop gracefully shuts down all server components.
// It stops the certificate renewal process, shuts down all servers,
// and closes all channels.
func (s *zephyrixServer) stop(ctx context.Context) error {
	s.stopCertRenewal()

	if err := s.stopChallengeServer(ctx); err != nil {
		s.logger.Error("Failed to stop challenge server %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(s.servers))

	for _, srv := range s.servers {
		wg.Add(1)
		go func(server *http.Server) {
			defer wg.Done()
			if err := server.Shutdown(ctx); err != nil {
				errChan <- fmt.Errorf("server shutdown error: %w", err)
			}
		}(srv)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var lastErr error
	for err := range errChan {
		s.logger.Error("Server shutdown error %w", err)
		lastErr = err
	}

	close(s.shutdownChan)
	close(s.errChannel)

	return lastErr
}

// spawnServer starts a specific server (HTTP or HTTPS).
// It sets up the handler and begins listening on the specified address.
func (s *zephyrixServer) spawnServer(serverType serverType, srv *http.Server) error {
	srv.Handler = s.z.setupHandler(s.handlers, s.middlewares)
	s.z.r.execute()

	var err error
	switch serverType {
	case serverHTTP:
		s.logger.Info("Starting HTTP server @ %s", srv.Addr)
		err = srv.ListenAndServe()
	case serverHTTPS:
		s.logger.Info("Starting HTTPS server @ %s", srv.Addr)
		err = srv.ListenAndServeTLS("", "") // Certificates are handled by the TLS config
	default:
		return fmt.Errorf("unknown server type: %v", serverType)
	}

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%v server error: %w", serverType, err)
	}
	return nil
}

// stopCertRenewal stops the certificate renewal process if it's running.
func (s *zephyrixServer) stopCertRenewal() {
	if s.certRenewTicker != nil {
		s.certRenewTicker.Stop()
		s.certRenewTicker = nil
	}
}
