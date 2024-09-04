package zephyrix

import "net/http"

// challengeServer encapsulates the HTTP server used for ACME challenges
type challengeServer struct {
	server *http.Server
	done   chan struct{}
}
