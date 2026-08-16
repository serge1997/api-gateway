package contracts

import "net/http"

type Route interface {
	Method() string
	Handler() http.HandlerFunc
	Path() string
}
type Router interface {
	Routes() []Route
}
