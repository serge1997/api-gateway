package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/serge1997/apigateway/internal/proxy"
	"github.com/serge1997/apigateway/internal/service"
)

func main() {
	proxy.Use("logger", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			log.Printf("%s - %s", r.Method, r.URL.Path)
			next(w, r)
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serviceName := r.Header.Get("x-service-name")
		srvce := service.Get(serviceName)
		if srvce == nil {
			fmt.Fprint(w, "service not found")
			return
		}
		prxy := proxy.New(srvce, r, w)
		prxy.Call(r.Context())
	})
	log.Fatal(http.ListenAndServe(":9091", nil))
}
