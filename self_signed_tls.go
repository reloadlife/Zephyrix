package zephyrix

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// isDevelopmentMode checks if the server is running in development mode
func (s *zephyrixServer) isDevelopmentMode() bool {
	return s.config.Environment == "development"
}

// generateSelfSignedCert generates a self-signed certificate for development purposes
func (s *zephyrixServer) generateSelfSignedCert() error {
	if !s.isDevelopmentMode() {
		return fmt.Errorf("self-signed certificates should only be used in development mode")
	}

	privateKey, err := s.generatePrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	template, err := s.createCertificateTemplate()
	if err != nil {
		return fmt.Errorf("failed to create certificate template: %w", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	if err := s.saveCertificateAndKey(certBytes, privateKey); err != nil {
		return fmt.Errorf("failed to save certificate and key: %w", err)
	}

	return nil
}

// generatePrivateKey creates a new ECDSA private key
func (s *zephyrixServer) generatePrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// createCertificateTemplate sets up the certificate template with appropriate values
func (s *zephyrixServer) createCertificateTemplate() (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Zephyrix Development"},
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	return template, nil
}

// saveCertificateAndKey writes the certificate and private key to disk
func (s *zephyrixServer) saveCertificateAndKey(certBytes []byte, privateKey *ecdsa.PrivateKey) error {
	certOut, err := os.Create("server.crt")
	if err != nil {
		return fmt.Errorf("failed to open server.crt for writing: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("failed to write data to server.crt: %w", err)
	}

	keyOut, err := os.OpenFile("server.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open server.key for writing: %w", err)
	}
	defer keyOut.Close()

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write data to server.key: %w", err)
	}

	return nil
}
