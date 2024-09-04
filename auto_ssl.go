package zephyrix

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// configureAutoSSL sets up automatic SSL using Let's Encrypt or ZeroSSL.
// It configures the HTTPS server with automatic certificate management.
//
// This function performs the following tasks:
// 1. Validates the presence of required configuration parameters.
// 2. Sets up an autocert.Manager for certificate management.
// 3. Configures ZeroSSL if specified in the configuration.
// 4. Creates and configures a TLS configuration.
// 5. Starts an HTTP-01 challenge server for domain validation.
// 6. Sets up the HTTPS server with the generated TLS configuration.
//
// Returns an error if any step in the configuration process fails.
func (s *zephyrixServer) configureAutoSSL() error {
	if err := s.validateAutoSSLConfig(); err != nil {
		return err
	}

	certManager := s.createCertManager()
	tlsConfig := s.createTLSConfig(certManager)

	if err := s.configureTLSOptions(tlsConfig); err != nil {
		return fmt.Errorf("failed to configure TLS options: %w", err)
	}

	challengeSrv, err := s.startChallengeServer(certManager)
	if err != nil {
		return fmt.Errorf("failed to start challenge server: %w", err)
	}
	s.challengeServer = challengeSrv
	return nil
}

// validateAutoSSLConfig checks if the necessary configuration for AutoSSL is present.
func (s *zephyrixServer) validateAutoSSLConfig() error {
	if len(s.config.SSL.AutoSSLDomains) == 0 {
		return fmt.Errorf("auto_ssl_domains must be specified when auto_ssl is enabled")
	}

	if s.config.SSL.AutoSSLProvider == "zerossl" {
		if s.config.SSL.AutoSSLZeroSSLEABKey == "" || s.config.SSL.AutoSSLZeroSSLKID == "" {
			return fmt.Errorf("ZeroSSL EAB key and KID must be provided when using ZeroSSL")
		}
	}

	return nil
}

// createCertManager initializes and returns an autocert.Manager based on the server configuration.
func (s *zephyrixServer) createCertManager() *autocert.Manager {
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.SSL.AutoSSLDomains...),
		Email:      s.config.SSL.AutoSSLEmail,
	}

	if s.config.SSL.AutoSSLCacheDir != "" {
		certManager.Cache = autocert.DirCache(s.config.SSL.AutoSSLCacheDir)
	}

	if s.config.SSL.AutoSSLProvider == "zerossl" {
		certManager.Client = &acme.Client{
			DirectoryURL: "https://acme.zerossl.com/v2/DV90",
		}
		certManager.ExternalAccountBinding = &acme.ExternalAccountBinding{
			KID: s.config.SSL.AutoSSLZeroSSLKID,
			Key: []byte(s.config.SSL.AutoSSLZeroSSLEABKey),
		}
	}

	return certManager
}

// createTLSConfig creates a TLS configuration using the provided cert manager.
func (s *zephyrixServer) createTLSConfig(certManager *autocert.Manager) *tls.Config {
	return &tls.Config{
		GetCertificate: certManager.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1"},
	}
}

// startChallengeServer starts an HTTP server to handle ACME HTTP-01 challenges.
func (s *zephyrixServer) startChallengeServer(certManager *autocert.Manager) (*challengeServer, error) {
	srv := &http.Server{
		Addr:    s.config.SSL.AutoSSLChallangeAddr,
		Handler: certManager.HTTPHandler(nil),
	}

	challengeSrv := &challengeServer{
		server: srv,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(challengeSrv.done)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.errChannel <- fmt.Errorf("challenge server error: %w", err)
		}
	}()

	// Wait for the server to start or fail
	select {
	case <-time.After(5 * time.Second):
		return challengeSrv, nil
	case <-challengeSrv.done:
		return nil, fmt.Errorf("challenge server failed to start")
	}
}

// stopChallengeServer gracefully shuts down the challenge server.
func (s *zephyrixServer) stopChallengeServer(ctx context.Context) error {
	if s.challengeServer == nil {
		return nil
	}

	shutdownErr := s.challengeServer.server.Shutdown(ctx)

	// Wait for the server to finish or the context to be canceled
	select {
	case <-s.challengeServer.done:
	case <-ctx.Done():
		return fmt.Errorf("challenge server shutdown timed out: %w", ctx.Err())
	}

	return shutdownErr
}
