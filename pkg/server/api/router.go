package api

import (
	"github.com/gorilla/mux"
)

// NewRouter creates a new router with configured routes
func NewRouter() *mux.Router {
	return mux.NewRouter()
}
