package retry

import "github.com/serge1997/apigateway/internal/service"

// service need to import retry, move WithRetry to proxy package
func WithRetry(srvce service.Service, caller func() error) {

}
