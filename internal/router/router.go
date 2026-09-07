package router

import (
	"net/http"
	"slices"

	"github.com/serge1997/apigateway/internal/contracts"
)

type Route struct {
	path    string
	methods []string
	hanlder http.HandlerFunc
}
type router struct {
	routes []contracts.Route
}

func (r Route) Handler() http.HandlerFunc {
	return r.hanlder
}
func (r Route) Methods() []string {
	return r.methods
}

func (r Route) Path() string {
	return r.path
}
func (r router) Routes() []contracts.Route {
	return r.routes
}

func New() router {
	var rts = make([]contracts.Route, len(routes))
	for _, route := range routes {
		if !slices.Contains(route.Methods(), http.MethodOptions) {
			route.methods = append(route.methods, http.MethodOptions)
		}
		rts = append(rts, route)
	}
	return router{routes: rts}
}
