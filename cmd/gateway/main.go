package main

import (
	"log"
	"net/http"

	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/proxy"
)

func main() {
	gtw := apigateway.New(apigateway.APIGatewayConfig{})
	gtw.UseGlobal("logger", func(ctx *proxy.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("%s - %s", ctx.Method(), ctx.Path())
		return next, nil
	})
	gtw.Use("auth_jwt", func(ctx *proxy.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		token := ctx.Req().Header.Get("Authorization")
		if token == "" {
			return next, ctx.Unauthorized()
		}
		return next, nil
	})
	log.Fatal(gtw.Listen())
}
