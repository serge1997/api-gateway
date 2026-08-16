package apigateway

import (
	"time"

	"github.com/serge1997/apigateway/internal/contracts"
)

var cORSAllowedOrigins []string

type APIGatewayConfig struct {
	ListenAddr         string        `yaml:"listen_addr" json:"listenAddr"`
	ReadTimeout        time.Duration `yaml:"read_timeout" json:"readTimeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout" json:"writeTimeout"`
	Timeout            time.Duration `yaml:"timeout" json:"timeout"`
	CORSAllowedOrigins []string      `yaml:"cors_allowed_origins" json:"corsAllowedOrigin"`
}

type Config struct {
	Routes []contracts.Router
}
