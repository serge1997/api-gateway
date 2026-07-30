package router

import "net/http"

type Router struct {
	Path    string
	Method  string
	Hanlder http.HandlerFunc
}

func Routes() []Router {
	return routes
}
