package apigateway

import "time"

type APIGatewayConfig struct {
	ListenAddr        string        `yaml:"listen_addr" json:"listenAddr"`
	ReadTimeout       time.Duration `yaml:"read_timeout" json:"readTimeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" json:"writeTimeout"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
	CORSAllowedOrigin []string      `yaml:"cors_allowed_origin" json:"corsAllowedOrigin"`
}
