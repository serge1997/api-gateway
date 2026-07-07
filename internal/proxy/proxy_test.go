package proxy_test

import (
	"net/http"
	"testing"

	"github.com/serge1997/apigateway/internal/proxy"
)

func TestMiddlewareHasBeenAdded(t *testing.T) {
	proxy.Use("testing", func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			hf(w, r)
		}
	})
	if len(proxy.Middlewares()) == 0 {
		t.Errorf("[middleware not added] expected %v got %v", len(proxy.Middlewares()) > 0, len(proxy.Middlewares()) == 0)
	}
}

func TestMiddlewaresMapHasKey(t *testing.T) {
	key := "testing"
	_, ok := proxy.Middlewares()[key]
	if ok == false {
		t.Errorf("[find middleware by key] expected find middleware by key %v got %v", key, ok)
	}
}
