package contracts

import "net/http"

type Route interface {
	Methods() []string
	Handler() http.HandlerFunc
	Path() string
}
type Router interface {
	Routes() []Route
}
