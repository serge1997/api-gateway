package router

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/api"
)

var routes = []Route{
	{
		path:    "/",
		methods: []string{http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch, http.MethodPost},
		hanlder: api.ServicesHandler,
	},
}
