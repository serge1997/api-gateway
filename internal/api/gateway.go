package api

import (
	"fmt"
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/middleware"
	"github.com/serge1997/apigateway/internal/middleware/cors"

	"github.com/serge1997/apigateway/internal/proxy"
)

func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	if cors.Handler(w, r, apigateway.Cors()) {
		w.WriteHeader(http.StatusOK)
		return
	}
	serviceName := apigateway.ExtractServiceName(r)
	srvce := apigateway.GetService(serviceName)
	if srvce == nil {
		http.Error(w, proxy.Err(fmt.Errorf("service %s not found", serviceName), http.StatusNotFound).Error(), http.StatusNotFound)
		return
	}
	combindedMdlws := middleware.MiddlewaresOf(srvce)
	prxy := proxy.New(srvce, combindedMdlws, w, r)
	prxy.Call(r.Context())
}
