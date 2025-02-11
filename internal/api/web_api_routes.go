package api

import (
	"net/http"

	"github.com/bendigiorgio/go-kv/internal/utils"
	"github.com/bendigiorgio/go-kv/internal/web/views/components"
)

type WebApiRouteHandler func(w http.ResponseWriter, req *http.Request) error

func (r *Router) handleRefreshList(w http.ResponseWriter, req *http.Request) error {
	limit := utils.StringToInt(req.URL.Query().Get("limit"), 50)
	page := utils.StringToInt(req.URL.Query().Get("page"), 1)
	var kvPairs = r.store.GetSlice(limit, page)
	return components.ListInner(kvPairs, page, limit).Render(req.Context(), w)
}

func (r *Router) handleDashboardStats(w http.ResponseWriter, req *http.Request) error {
	keyCount := r.store.KeyCount()
	bytesUsage := r.store.MemoryUsage()
	bytesMemoryLimit := r.store.GetMemoryLimit()
	return components.DashboardTitleInner(keyCount, bytesUsage, bytesMemoryLimit).Render(req.Context(), w)
}
