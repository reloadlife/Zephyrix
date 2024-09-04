package zephyrix

type HTTPVerb string

const (
	GET     HTTPVerb = "GET"
	POST    HTTPVerb = "POST"
	PUT     HTTPVerb = "PUT"
	DELETE  HTTPVerb = "DELETE"
	PATCH   HTTPVerb = "PATCH"
	OPTIONS HTTPVerb = "OPTIONS"
	HEAD    HTTPVerb = "HEAD"
	CONNECT HTTPVerb = "CONNECT"
	TRACE   HTTPVerb = "TRACE"
)

type serverType int

const (
	serverHTTP serverType = iota
	serverHTTPS

	// serverGRPC   // ? not implemented yet
)
