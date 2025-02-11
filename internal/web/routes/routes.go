package routes

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bendigiorgio/go-kv/internal/engine"
	"github.com/bendigiorgio/go-kv/internal/utils"
	"github.com/bendigiorgio/go-kv/internal/web/views"
)

type Route struct {
	Title     string
	Component func(eng *engine.Engine, req *http.Request) templ.Component
	Path      string
}

func (r Route) Handle(w http.ResponseWriter, req *http.Request, eng *engine.Engine) error {
	innerPage := func() templ.Component {
		return r.Component(eng, req)
	}()
	page := views.Base(r.Title, innerPage)
	return page.Render(req.Context(), w)
}

var HomeRoute = Route{
	Title: "KV Dashboard",
	Component: func(eng *engine.Engine, _ *http.Request) templ.Component {
		return views.Home(eng)
	},
	Path: "",
}

var ListRoute = Route{
	Title: "Key Value List",
	Component: func(eng *engine.Engine, req *http.Request) templ.Component {
		page := req.URL.Query().Get("page")
		if page == "" {
			page = "1"
			req.URL.Query().Set("page", page)
		}
		limit := req.URL.Query().Get("limit")
		if limit == "" {
			limit = "50"
			req.URL.Query().Set("limit", limit)
		}

		var params views.ListQueryParams = views.ListQueryParams{
			Page:  utils.StringToInt(page, 1),
			Limit: utils.StringToInt(limit, 50),
		}

		return views.List(eng, params)
	},
	Path: "/list",
}

func GetRoutes() []Route {
	return []Route{HomeRoute, ListRoute}
}
