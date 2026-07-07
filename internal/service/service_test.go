package service_test

import (
	"testing"

	"github.com/serge1997/apigateway/internal/service"
)

func TestServiceHasBeenLoaded(t *testing.T) {
	srvcs := service.Services()
	if len(srvcs) == 0 {
		t.Errorf("expected service has been loaded, go %v", len(srvcs))
	}
}

func TestAServiceExists(t *testing.T) {
	expected := service.Get("users")
	if expected == nil {
		t.Errorf("expected instance of Service got %v", expected)
	}
}
