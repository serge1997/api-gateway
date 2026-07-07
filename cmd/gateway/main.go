package main

import (
	"log"
	"net/http"

	"github.com/serge1997/apigateway/internal/proxy"
)

func main() {
	proxy.Use("logger", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			log.Printf("%s - %s", r.Method, r.URL.Path)
			next(w, r)
		}
	})
	proxy.Use("auth_jwt", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	})
	proxy.Use("cors", func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next(w, r)
		}
	})
	log.Fatal(http.ListenAndServe(":9091", proxy.Router()))
}
