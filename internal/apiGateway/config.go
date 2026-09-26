package apigateway

import (
	"time"

	"github.com/serge1997/apigateway/internal/contracts"
)

var defaultServiceLookupHeader = "x-gateway-service"
var cORSAllowedOrigins []string

type APIGatewayConfig struct {
	ListenAddr         string        `yaml:"listen_addr" json:"listenAddr"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	CORSAllowedOrigins []string      `yaml:"cors_allowed_origins" json:"corsAllowedOrigin"`
	ReadHeaderTimeout  time.Duration `yaml:"read_header_timeout" json:"readHeaderTimeout"`
}

type Config struct {
	Routes []contracts.Router
}
