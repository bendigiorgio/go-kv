package api

import (
	"net/http"

	"github.com/bendigiorgio/go-kv/internal/web/routes"
)

func (r *Router) getWebRouteHandler(route routes.Route) http.HandlerFunc {
	return routes.NewHandler(routes.CustomHandler(route.Handle), r.store)
}

func (r *Router) wrapWebApiRouteHandler(routeHandler WebApiRouteHandler) http.HandlerFunc {
	return routes.NewHandlerNoEngine(routes.CustomHandlerNoEngine(routeHandler))
}
