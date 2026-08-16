package router

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/api"
)

var routes = []Route{
	{
		path:   "/health",
		method: http.MethodGet,
		hanlder: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("200 OK - healthy\n"))
		},
	},
	{
		path:    "/",
		method:  http.MethodGet,
		hanlder: api.ServicesHandler,
	},
}
