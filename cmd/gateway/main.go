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
	srv := server.New(gtw, router.New())
	gtw.UseGlobal("logger", func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		log.Printf("%s - %s", ctx.Method(), ctx.Path())
		return next, nil
	})
	log.Fatal(srv.Listen())
}
