package manager

import "net/http"

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	return mux
}
