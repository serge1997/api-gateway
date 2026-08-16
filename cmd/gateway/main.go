package main

import (
	"log"
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/proxy"
	"github.com/serge1997/apigateway/internal/router"
	"github.com/serge1997/apigateway/internal/server"
)

func main() {
	gtw := apigateway.New(apigateway.Config{})
	srv := server.New(gtw, router.New())

	gtw.UseGlobal("logger", func(ctx *proxy.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("%s - %s", ctx.Method(), ctx.Path())
		return next, nil
	})
	gtw.UseGlobal("permission", func(ctx *proxy.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		//return nil, ctx.Err(fmt.Errorf("permission denied"), 403)
		return next, nil
	})
	gtw.Use("auth_jwt", func(ctx *proxy.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		token := ctx.Req().Header.Get("Authorization")
		if token == "" {
			return nil, ctx.Unauthorized()
		}
		return next, nil
	})
	log.Fatal(srv.Listen())
}
