package apigateway

import (
	"io"
	"net/http"
)

type Context interface {
	Req() *http.Request
	Writer() http.ResponseWriter
	Path() string
	Method() string
	Body() io.Reader
	Unauthorized() error
	Err(err error, status int) error
}
