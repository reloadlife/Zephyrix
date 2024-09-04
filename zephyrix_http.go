package zephyrix

import (
	"net/http"
)

func (s *zephyrixServer) configureHTTPServer() error {
	s.servers[serverHTTP] = &http.Server{
		Addr:         s.config.Address,
		ReadTimeout:  s.config.ParsedReadTimeout,
		WriteTimeout: s.config.ParsedWriteTimeout,
		IdleTimeout:  s.config.ParsedIdleTimeout,
	}
	return nil
}

func (s *zephyrixServer) SetupHTTPRedirect() {
	httpServer, ok := s.servers[serverHTTP]
	if !ok {
		s.logger.Warn("HTTP server not configured, cannot set up redirect")
		return
	}

	if s.config.RedirectToHTTPS {
		httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})
	}
}
