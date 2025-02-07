package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bendigiorgio/go-kv/internal/engine"
	internal "github.com/bendigiorgio/go-kv/internal/web"
	"github.com/bendigiorgio/go-kv/internal/web/routes"
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
	r.mux.HandleFunc("/set", r.handleSet)
	r.mux.HandleFunc("/get", r.handleGet)
	r.mux.HandleFunc("/delete", r.handleDelete)
	r.mux.HandleFunc("/list", r.handleList)
	r.mux.HandleFunc("/flush", r.handleFlush)
	r.mux.HandleFunc("/compact", r.handleCompact)
	r.mux.HandleFunc("/memory-usage", r.handleGetMemoryUsage)
	r.mux.HandleFunc("/count", r.handleGetKeyCount)
	r.mux.HandleFunc("/batch/set", r.handleBatchSet)
	r.mux.HandleFunc("/batch/delete", r.handleBatchDelete)

	// WEB UI
	if useWebUI {
		var web_routes = routes.GetRoutes()
		fs := http.FileServer(http.FS(internal.StaticFiles))
		r.mux.Handle("/web/static/", http.StripPrefix("/web/", fs))
		for _, route := range web_routes {
			var handler = r.getWebRouteHandler(route)
			var path = "/web" + route.Path
			r.mux.HandleFunc(path, handler)
		}
		r.mux.HandleFunc("/web/*", func(w http.ResponseWriter, req *http.Request) {
			http.NotFound(w, req)
		})

		r.mux.HandleFunc("/web/api/list", r.wrapWebApiRouteHandler(r.handleRefreshList))
		r.mux.HandleFunc("/web/api/dashboard", r.wrapWebApiRouteHandler(r.handleDashboardStats))
	}
}

// Start runs the HTTP server on the specified port
func (r *Router) Start(port string) error {
	r.server = &http.Server{
		Addr:    ":" + port,
		Handler: r.mux,
	}

	log.Printf("Server starting on port %s\n", port)
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

	log.Println("Shutting down server...")
	return r.server.Shutdown(ctx)
}

// Respond with JSON helper
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
