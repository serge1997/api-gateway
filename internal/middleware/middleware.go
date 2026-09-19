package middleware

import (
	"sync"

	"github.com/serge1997/apigateway/internal/proxy"
	"github.com/serge1997/apigateway/internal/service"
)

type Handler = proxy.MiddlewareHandler
type MiddlewareMap = proxy.MiddlewareHandlerMap

var middlewares MiddlewareMap = map[string]Handler{}
var globales MiddlewareMap = map[string]Handler{}
var combined proxy.CombinedMiddlewares = map[string]proxy.MiddlewareHandlerMap{}
var cache = map[string]MiddlewareMap{}
var mu sync.RWMutex

func Register(name string, handler Handler) {
	_, ok := middlewares[name]
	if ok == false {
		middlewares[name] = handler
	}
}
func RegisterGlobal(name string, handler Handler) {
	_, ok := globales[name]
	if ok == false {
		globales[name] = handler
	}
}

func Middlewares() map[string]Handler {
	return middlewares
}

func Globales() map[string]Handler {
	return globales
}

func MiddlewaresOf(s *service.Service) MiddlewareMap {
	mu.RLock()
	serviceMdlws, ok := cache[s.Name]
	if ok {
		mu.RUnlock()
		return serviceMdlws
	}
	mu.RUnlock()
	mu.Lock()
	defer mu.Unlock()
	mdlws := MiddlewareMap{}
	for _, mdlwName := range s.Middlewares {
		h, ok := middlewares[mdlwName]
		if ok {
			mdlws[mdlwName] = h
		}
	}
	cache[s.Name] = mdlws
	return mdlws
}

func CombinedGlobalWithService(s *service.Service) proxy.CombinedMiddlewares {
	serviceMdlws := MiddlewaresOf(s)
	combined := proxy.CombinedMiddlewares{}
	combined["service"] = serviceMdlws
	combined["global"] = globales
	return combined
}
