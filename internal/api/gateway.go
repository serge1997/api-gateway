package api

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/middleware"

	"github.com/serge1997/apigateway/internal/proxy"
	"github.com/serge1997/apigateway/internal/service"
)

func ServicesHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.Header.Get("x-service-name")
	srvce := service.Get(serviceName)
	if srvce == nil {
		http.Error(w, "service not found", 404)
		return
	}
	combindedMdlws := middleware.CombinedGlobalWithService(srvce)
	prxy := proxy.New(srvce, combindedMdlws, w, r)
	prxy.Call(r.Context())
}
