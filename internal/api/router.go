package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bendigiorgio/go-kv/internal/engine"
	"github.com/bendigiorgio/go-kv/internal/web/routes"
	"github.com/rs/zerolog/log"

	internal "github.com/bendigiorgio/go-kv/internal/web"
)

// Router is a simple HTTP router with graceful shutdown and an Engine reference
type Router struct {
	mux    *http.ServeMux
	server *http.Server
	store  *engine.Engine
}

// NewRouter initializes a new Router with a key-value store
func NewRouter(store *engine.Engine, useWebUI bool) *Router {
	r := &Router{
		mux:   http.NewServeMux(),
		store: store,
	}
	r.registerRoutes(useWebUI)
	return r
}

// ServeHTTP makes Router satisfy the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// registerRoutes sets up API endpoints
func (r *Router) registerRoutes(useWebUI bool) {
	// API Routes
	apiRoutes := map[string]http.HandlerFunc{
		"/set":               r.handleSet,
		"/get":               r.handleGet,
		"/delete":            r.handleDelete,
		"/list":              r.handleList,
		"/flush":             r.handleFlush,
		"/compact":           r.handleCompact,
		"/memory-usage":      r.handleGetMemoryUsage,
		"/count":             r.handleGetKeyCount,
		"/batch/set":         r.handleBatchSet,
		"/batch/delete":      r.handleBatchDelete,
		"/web/api/list":      r.wrapWebApiRouteHandler(r.handleRefreshList),
		"/web/api/dashboard": r.wrapWebApiRouteHandler(r.handleDashboardStats),
	}

	for path, handler := range apiRoutes {
		r.mux.HandleFunc(path, handler)
	}

	// WEB UI
	if useWebUI {
		fs := http.FileServer(http.FS(internal.StaticFiles))
		r.mux.Handle("/web/static/", http.StripPrefix("/web/", fs))

		for _, route := range routes.GetRoutes() {
			r.mux.HandleFunc("/web"+route.Path, r.getWebRouteHandler(route))
		}

		r.mux.HandleFunc("/web/*", func(w http.ResponseWriter, req *http.Request) {
			http.NotFound(w, req)
		})
	}
}

// Start runs the HTTP server on the specified port
func (r *Router) Start(port string) error {
	r.server = &http.Server{
		Addr:    ":" + port,
		Handler: r.mux,
	}

	log.Info().Msgf("Server starting on port %s\n", port)
	err := r.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop gracefully shuts down the server
func (r *Router) Stop() error {
	if r.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down server...")
	return r.server.Shutdown(ctx)
}

// Respond with JSON helper
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Stack().Err(err).Msg("Failed to encode JSON response")
	}
}
