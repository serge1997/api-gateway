package api

import (
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/middleware"

	"github.com/serge1997/apigateway/internal/proxy"
)

func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.Header.Get("x-service-name")
	srvce := apigateway.GetService(serviceName)
	if srvce == nil {
		http.Error(w, "service not found", 404)
		return
	}
	combindedMdlws := middleware.CombinedGlobalWithService(srvce)
	prxy := proxy.New(srvce, combindedMdlws, w, r)
	prxy.Call(r.Context())
}
