package routes

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bendigiorgio/go-kv/internal/engine"
	"github.com/bendigiorgio/go-kv/internal/web/views"
)

type Route struct {
	Title     string
	Component func(*engine.Engine) templ.Component
	Path      string
}

func (r Route) Handle(w http.ResponseWriter, req *http.Request, eng *engine.Engine) error {
	innerPage := func() templ.Component {
		return r.Component(eng)
	}()
	page := views.Base(r.Title, innerPage)
	return page.Render(req.Context(), w)
}

// Define HomeRoute without manually using Layout
var HomeRoute = Route{
	Title: "KV Dashboard",
	Component: func(eng *engine.Engine) templ.Component {
		return views.Home(eng) // No need to manually wrap in Layout anymore!
	},
	Path: "",
}

var ListRoute = Route{
	Title: "Key Value List",
	Component: func(eng *engine.Engine) templ.Component {
		return views.List(eng)
	},
	Path: "/list",
}

// GetRoutes returns all defined routes
func GetRoutes() []Route {
	return []Route{HomeRoute, ListRoute}
}
