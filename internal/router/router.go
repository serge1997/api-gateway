package router

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/contracts"
)

type Route struct {
	path    string
	method  string
	hanlder http.HandlerFunc
}
type router struct {
	routes []contracts.Route
}

func (r Route) Handler() http.HandlerFunc {
	return r.hanlder
}
func (r Route) Method() string {
	return r.method
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
		routeOption := route
		routeOption.method = http.MethodOptions
		rts = append(rts, route, routeOption)
	}
	return router{routes: rts}
}
