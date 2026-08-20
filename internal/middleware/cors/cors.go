package cors

import (
	"net/http"

	"github.com/serge1997/apigateway/internal/contracts"
)

func New() func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
	return func(ctx contracts.Context, next http.HandlerFunc) (http.HandlerFunc, error) {
		ctx.Writer().Header().Set("Access-Control-Allow-Origin", "*")
		return next, nil
	}
}
