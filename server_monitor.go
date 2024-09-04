package zephyrix

import (
	"context"
	"time"
)

func (s *zephyrixServer) monitorErrors(ctx context.Context) {
	for {
		select {
		case err, ok := <-s.errChannel:
			if !ok {
				return
			}
			s.logger.Error("Server error", "error", err)
			if err := s.z.Stop(); err != nil {
				s.logger.Fatal("Failed to stop the server", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *zephyrixServer) startCertificateRenewalMonitor(ctx context.Context) {
	s.certRenewTicker = time.NewTicker(24 * time.Hour)
	defer s.certRenewTicker.Stop()

	for {
		select {
		case <-s.certRenewTicker.C:
			if err := s.renewCertificatesIfNeeded(); err != nil {
				s.logger.Error("Failed to renew certificates", "error", err)
			}
		case <-s.shutdownChan:
			return
		case <-ctx.Done():
			return
		}
	}
}
