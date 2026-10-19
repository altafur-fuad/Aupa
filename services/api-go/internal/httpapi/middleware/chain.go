package middleware

import "net/http"

// Middleware defines a standard HTTP middleware wrapper function.
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middlewares into a single http.Handler execution pipeline.
// Middlewares are executed in the order they are passed.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
