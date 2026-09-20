package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
)

type Context struct {
	req    *http.Request
	writer http.ResponseWriter
}

func (c *Context) Req() *http.Request {
	return c.req
}
func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Context) Path() string {
	return c.req.URL.Path
}
func (c *Context) Method() string {
	return c.req.Method
}
func (c *Context) Body() io.Reader {
	return c.req.Body
}

func (c *Context) Unauthorized() error {
	return fmt.Errorf("%s", httpresponse.ToJSON(
		"Unauthorized",
		http.StatusUnauthorized,
		false,
		"",
		//string(debug.Stack()),
	))
}
func (c *Context) Err(err error, status int) error {
	return Err(err, status)
}

func (c *Context) Host() string {
	return c.req.Host
}

func (c *Context) Header(key string) string {
	return c.req.Header.Get(key)
}

func (c *Context) GetStatusString(str string) int {
	var status = 501
	var errMap map[string]any
	if err := json.Unmarshal([]byte(str), &errMap); err != nil {
		return status
	}
	if value, ok := errMap["status"]; ok {
		status = int(value.(float64))
	}
	return status
}

func Err(err error, status int) error {
	return fmt.Errorf("%s", httpresponse.ToJSON(
		err.Error(),
		status,
		false,
		"",
		//string(debug.Stack()),
	))
}
