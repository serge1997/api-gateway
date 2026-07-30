package apigateway

import (
	"fmt"
	"net/http"

	"github.com/serge1997/apigateway/internal/middleware"
	"github.com/serge1997/apigateway/internal/router"
	"github.com/serge1997/apigateway/internal/service"
)

type apiGateway struct {
	Config   APIGatewayConfig            `yaml:"config" json:"config"`
	Services map[string]*service.Service `yaml:"service" json:"services"`
	mux      *http.ServeMux              `yaml:"-" json:"-"`
}

func New(conf APIGatewayConfig) *apiGateway {
	return &apiGateway{
		Config: conf,
		mux:    http.NewServeMux(),
	}
}

func (a *apiGateway) Listen() error {
	return http.ListenAndServe(a.Config.ListenAddr, a.serveHTTP())
}

func (a *apiGateway) serveHTTP() http.Handler {
	routes := router.Routes()
	for _, route := range routes {
		path := fmt.Sprintf("%s %s", route.Method, route.Path)
		a.mux.HandleFunc(path, route.Hanlder)
	}
	return a.mux
}

func (a *apiGateway) Use(name string, handler middleware.Handler) {
	middleware.Register(name, handler)
}

func (a *apiGateway) UseGlobal(name string, handler middleware.Handler) {
	middleware.RegisterGlobal(name, handler)
}
