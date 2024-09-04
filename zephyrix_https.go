package zephyrix

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

func (s *zephyrixServer) configureHTTPSServer() error {
	if !s.config.SSL.Enabled {
		return nil
	}

	if s.config.SSL.AutoSSL {
		return s.configureAutoSSL()
	}
	return s.configureManualSSL()
}

// configureManualSSL sets up SSL using provided certificate and key files
func (s *zephyrixServer) configureManualSSL() error {
	if s.isDevelopmentMode() {
		if s.config.SSL.CertFile == "" || s.config.SSL.KeyFile == "" {
			if err := s.generateSelfSignedCert(); err != nil {
				return fmt.Errorf("failed to create self-signed certificate: %w", err)
			}
			s.config.SSL.CertFile = "server.crt"
			s.config.SSL.KeyFile = "server.key"
		}
	}
	cert, err := tls.LoadX509KeyPair(s.config.SSL.CertFile, s.config.SSL.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if err := s.configureTLSOptions(tlsConfig); err != nil {
		return err
	}

	s.servers[serverHTTPS] = &http.Server{
		Addr:         s.config.SSL.Address,
		ReadTimeout:  s.config.ParsedReadTimeout,
		WriteTimeout: s.config.ParsedWriteTimeout,
		IdleTimeout:  s.config.ParsedIdleTimeout,
		TLSConfig:    tlsConfig,
	}

	return nil
}

// configureTLSOptions sets up various TLS options based on the configuration
func (s *zephyrixServer) configureTLSOptions(tlsConfig *tls.Config) error {
	if err := s.configureTLSVersions(tlsConfig); err != nil {
		return err
	}

	if err := s.configureTLSCipherSuites(tlsConfig); err != nil {
		return err
	}

	if err := s.configureTLSCurvePreferences(tlsConfig); err != nil {
		return err
	}

	if err := s.configureTLSClientAuth(tlsConfig); err != nil {
		return err
	}

	tlsConfig.Renegotiation = s.getTLSRenegotiationSupport()

	return nil
}

// configureTLSVersions sets the minimum and maximum TLS versions
func (s *zephyrixServer) configureTLSVersions(tlsConfig *tls.Config) error {
	minVersion, err := parseTLSVersion(s.config.SSL.MinVersion, tls.VersionTLS12)
	if err != nil {
		return fmt.Errorf("invalid TLS min version: %w", err)
	}

	maxVersion, err := parseTLSVersion(s.config.SSL.MaxVersion, tls.VersionTLS13)
	if err != nil {
		return fmt.Errorf("invalid TLS max version: %w", err)
	}

	if minVersion > maxVersion {
		return fmt.Errorf("TLS min version (%s) is greater than max version (%s)",
			s.config.SSL.MinVersion, s.config.SSL.MaxVersion)
	}

	tlsConfig.MinVersion = minVersion
	tlsConfig.MaxVersion = maxVersion
	return nil
}

// parseTLSVersion converts a string TLS version to its uint16 representation
func parseTLSVersion(version string, defaultVersion uint16) (uint16, error) {
	switch strings.ToLower(version) {
	case "1.0", "tls1.0":
		return tls.VersionTLS10, nil
	case "1.1", "tls1.1":
		return tls.VersionTLS11, nil
	case "1.2", "tls1.2":
		return tls.VersionTLS12, nil
	case "1.3", "tls1.3":
		return tls.VersionTLS13, nil
	case "":
		return defaultVersion, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

// configureTLSCipherSuites sets the allowed cipher suites
func (s *zephyrixServer) configureTLSCipherSuites(tlsConfig *tls.Config) error {
	if len(s.config.SSL.CipherSuites) > 0 {
		cipherSuites, err := parseCipherSuites(s.config.SSL.CipherSuites)
		if err != nil {
			return fmt.Errorf("invalid cipher suite configuration: %w", err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}
	return nil
}

// parseCipherSuites converts a list of cipher suite names to their uint16 representations
func parseCipherSuites(cipherNames []string) ([]uint16, error) {
	var cipherSuites []uint16
	for _, name := range cipherNames {
		cipher, err := getCipherSuite(name)
		if err != nil {
			return nil, err
		}
		cipherSuites = append(cipherSuites, cipher)
	}
	return cipherSuites, nil
}

// getCipherSuite converts a cipher suite name to its uint16 representation
func getCipherSuite(name string) (uint16, error) {
	cipherMap := map[string]uint16{
		"TLS_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"TLS_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}

	cipher, ok := cipherMap[name]
	if !ok {
		return 0, fmt.Errorf("unsupported cipher suite: %s", name)
	}
	return cipher, nil
}

// configureTLSCurvePreferences sets the preferred elliptic curves
func (s *zephyrixServer) configureTLSCurvePreferences(tlsConfig *tls.Config) error {
	if len(s.config.SSL.CurvePreferences) > 0 {
		curves, err := parseCurvePreferences(s.config.SSL.CurvePreferences)
		if err != nil {
			return fmt.Errorf("invalid curve preferences configuration: %w", err)
		}
		tlsConfig.CurvePreferences = curves
	}
	return nil
}

// parseCurvePreferences converts a list of curve names to their tls.CurveID representations
func parseCurvePreferences(curveNames []string) ([]tls.CurveID, error) {
	var curves []tls.CurveID
	for _, name := range curveNames {
		curve, err := getCurveID(name)
		if err != nil {
			return nil, err
		}
		curves = append(curves, curve)
	}
	return curves, nil
}

// getCurveID converts a curve name to its tls.CurveID representation
func getCurveID(name string) (tls.CurveID, error) {
	curveMap := map[string]tls.CurveID{
		"P256":   tls.CurveP256,
		"P384":   tls.CurveP384,
		"P521":   tls.CurveP521,
		"X25519": tls.X25519,
	}

	curve, ok := curveMap[name]
	if !ok {
		return 0, fmt.Errorf("unsupported curve: %s", name)
	}
	return curve, nil
}

// configureTLSClientAuth sets up client certificate authentication
func (s *zephyrixServer) configureTLSClientAuth(tlsConfig *tls.Config) error {
	clientAuth, err := parseClientAuthType(s.config.SSL.ClientAuth)
	if err != nil {
		return fmt.Errorf("invalid client auth configuration: %w", err)
	}
	tlsConfig.ClientAuth = clientAuth

	if clientAuth != tls.NoClientCert && s.config.SSL.ClientCACert != "" {
		caCert, err := os.ReadFile(s.config.SSL.ClientCACert)
		if err != nil {
			return fmt.Errorf("failed to read client CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to parse client CA cert")
		}

		tlsConfig.ClientCAs = caCertPool
	}

	return nil
}

// parseClientAuthType converts a string client auth type to its tls.ClientAuthType representation
func parseClientAuthType(authType string) (tls.ClientAuthType, error) {
	switch strings.ToLower(authType) {
	case "no_client_cert", "":
		return tls.NoClientCert, nil
	case "request_client_cert":
		return tls.RequestClientCert, nil
	case "require_any_client_cert":
		return tls.RequireAnyClientCert, nil
	case "verify_client_cert_if_given":
		return tls.VerifyClientCertIfGiven, nil
	case "require_and_verify_client_cert":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return tls.NoClientCert, fmt.Errorf("unsupported client auth type: %s", authType)
	}
}

// getTLSRenegotiationSupport returns the appropriate TLS renegotiation setting
func (s *zephyrixServer) getTLSRenegotiationSupport() tls.RenegotiationSupport {
	if s.config.SSL.Renegotiation {
		return tls.RenegotiateOnceAsClient
	}
	return tls.RenegotiateNever
}

// ValidateSSLConfig checks the SSL configuration for consistency and completeness
func (s *zephyrixServer) ValidateSSLConfig() error {
	if !s.config.SSL.Enabled {
		return nil
	}

	if s.config.SSL.AutoSSL {
		if len(s.config.SSL.AutoSSLDomains) == 0 {
			return fmt.Errorf("auto_ssl_domains must be specified when auto_ssl is enabled")
		}
		if s.config.SSL.AutoSSLEmail == "" {
			return fmt.Errorf("auto_ssl_email must be specified when auto_ssl is enabled")
		}
	} else {
		if s.config.SSL.CertFile == "" || s.config.SSL.KeyFile == "" {
			return fmt.Errorf("cert_file and key_file must be specified when manual SSL is enabled")
		}
	}

	return nil
}

// LoadOrCreateTLSConfig loads an existing TLS configuration or creates a new one
func (s *zephyrixServer) LoadOrCreateTLSConfig() (*tls.Config, error) {
	if s.config.SSL.AutoSSL {
		return s.createAutoSSLConfig()
	}
	return s.createManualSSLConfig()
}

// createAutoSSLConfig creates a TLS configuration for automatic SSL
func (s *zephyrixServer) createAutoSSLConfig() (*tls.Config, error) {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.SSL.AutoSSLDomains...),
		Email:      s.config.SSL.AutoSSLEmail,
	}

	if s.config.SSL.AutoSSLCacheDir != "" {
		certManager.Cache = autocert.DirCache(s.config.SSL.AutoSSLCacheDir)
	}

	return &tls.Config{
		GetCertificate: certManager.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1"},
	}, nil
}

// createManualSSLConfig creates a TLS configuration for manual SSL
func (s *zephyrixServer) createManualSSLConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(s.config.SSL.CertFile, s.config.SSL.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// HandleCertificateError logs certificate-related errors and takes appropriate action
func (s *zephyrixServer) HandleCertificateError(err error) {
	Logger.Error("Certificate error", "error", err)
	// Depending on the error, you might want to trigger a certificate renewal,
	// notify an administrator, or take other corrective actions.

	//// TODO: Implement certificate error handling
}

// IsSSLEnabled returns whether SSL is enabled for the server
func (s *zephyrixServer) IsSSLEnabled() bool {
	return s.config.SSL.Enabled
}

// GetSSLPort returns the port number for SSL connections
func (s *zephyrixServer) GetSSLPort() string {
	if s.config.SSL.Address == "" {
		return "443" // Default HTTPS port
	}
	return strings.TrimPrefix(s.config.SSL.Address, ":")
}

// LogSSLConfig logs the current SSL configuration for debugging purposes
func (s *zephyrixServer) LogSSLConfig() {
	Logger.Debug("SSL Configuration %s: %t, %s: %t, %s: %s, %s: %s, %s: %s",
		"enabled", s.config.SSL.Enabled,
		"auto_ssl", s.config.SSL.AutoSSL,
		"address", s.config.SSL.Address,
		"min_version", s.config.SSL.MinVersion,
		"max_version", s.config.SSL.MaxVersion)
}
