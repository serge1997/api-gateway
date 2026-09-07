package proxy_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/serge1997/apigateway/internal/contracts"
	"github.com/serge1997/apigateway/internal/proxy"
)

var combinedMdlws proxy.CombinedMiddlewares = proxy.CombinedMiddlewares{
	"global": proxy.MiddlewareHandlerMap{
		"logger": func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
			fmt.Println("global middleware running")
			return next, nil
		},
	},
}

func TestMiddlewaresMapHasKey(t *testing.T) {
	key := "cors"
	_, ok := proxy.Middlewares()[key]
	if ok == false {
		t.Errorf("[find middleware by key] expected find middleware by key %v got %v", key, ok)
	}
}
