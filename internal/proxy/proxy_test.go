package proxy_test

import (
	"testing"

	"github.com/serge1997/apigateway/internal/proxy"
)

func TestMiddlewaresMapHasKey(t *testing.T) {
	key := "testing"
	_, ok := proxy.Middlewares()[key]
	if ok == false {
		t.Errorf("[find middleware by key] expected find middleware by key %v got %v", key, ok)
	}
}
