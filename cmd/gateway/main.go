package main

import (
	"log"
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/contracts"
	"github.com/serge1997/apigateway/internal/router"
	"github.com/serge1997/apigateway/internal/server"
)

func main() {
	gtw := apigateway.New()
	srv := server.New(
		server.WithApiGateway(gtw),
		server.WithRouter(router.New()),
	)
	defer srv.Close()
	gtw.UseGlobal("logger", func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("[logger] %s - %s", ctx.Method(), ctx.Path())
		return next, nil
	})
	gtw.Use("auth", func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		if ctx.Header("Authorization") != "Bearer my-secret-token" {
			return nil, ctx.Unauthorized()
		}
		return next, nil
	})
	log.Fatal(srv.Listen())
}
