package api

import (
	"net/http"

	"github.com/bendigiorgio/go-kv/internal/web/views"
	"github.com/bendigiorgio/go-kv/internal/web/views/components"
)

type WebApiRouteHandler func(w http.ResponseWriter, req *http.Request) error

func (r *Router) handleRefreshList(w http.ResponseWriter, req *http.Request) error {
	var kvPairs = views.GetListAsSlice(r.store)
	return components.ListInner(kvPairs).Render(req.Context(), w)
}

func (r *Router) handleDashboardStats(w http.ResponseWriter, req *http.Request) error {
	keyCount := r.store.KeyCount()
	bytesUsage := r.store.MemoryUsage()
	bytesMemoryLimit := r.store.GetMemoryLimit()
	return components.DashboardTitleInner(keyCount, bytesUsage, bytesMemoryLimit).Render(req.Context(), w)
}
