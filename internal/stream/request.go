package stream

import (
	"fmt"
	"time"

	"github.com/serge1997/apigateway/internal/database"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
	"github.com/serge1997/apigateway/shared/result"
	"gorm.io/gorm"
)

type requestGuard uint

var (
	RateLimitGuard      requestGuard = 1
	CircuitBreakerGuard requestGuard = 2
)

func init() {
	if err := database.Db().AutoMigrate(&RequestStream{}); err != nil {
		panic(err)
	}
	fmt.Println("migration done !")
}

type RequestStream struct {
	gorm.Model
	ID        int64        `json:"id"`
	Service   string       `json:"service"`
	Path      string       `json:"path"`
	Method    string       `json:"method"`
	Host      string       `json:"host"`
	Status    int          `json:"status"`
	Guard     requestGuard `json:"guard"`
	GuardType string       `json:"guardType"`
	Duration  int64        `json:"duration"` //millisecs
	Headers   string       `json:"headers"`
	CreatedAt time.Time    `json:"createdAt"`
	Response  string       `json:"response"`
}

func NewRequestStream(proxy WithProxyContract, response result.Result[httpresponse.HttpResponse]) *RequestStream {
	return &RequestStream{
		Service:   proxy.ServiceName(),
		Method:    proxy.Method(),
		Path:      proxy.Path(),
		Response:  response.Value().Message,
		Status:    response.Value().Status,
		CreatedAt: time.Now(),
		Host:      proxy.Host(),
		Headers:   proxy.HeadersToJson(),
		Duration:  proxy.Duration().Milliseconds(),
	}
}
