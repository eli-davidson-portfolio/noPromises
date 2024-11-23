package api

import (
	"net/http"
)

// API represents the main API server
type API struct {
	db     DB
	router http.Handler
}

// DB interface defines required database operations
type DB interface {
	// Add database methods as needed
}

// Option configures the API server
type Option func(*API)

// WithDB sets the database for the API
func WithDB(db DB) Option {
	return func(a *API) {
		a.db = db
	}
}

// New creates a new API server
func New(opts ...Option) *API {
	api := &API{}

	// Apply options
	for _, opt := range opts {
		opt(api)
	}

	return api
}

// ServeHTTP implements http.Handler
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.router != nil {
		a.router.ServeHTTP(w, r)
		return
	}

	// Default handler if no router is set
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
