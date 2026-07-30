package router

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/api"
)

var routes = []Router{
	{
		Path:    "/",
		Method:  http.MethodGet,
		Hanlder: api.ServicesHandler,
	},
}
