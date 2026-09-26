package contracts

import (
	"time"
)

type APIGateway interface {
	ListenAddr() string
	ReadHeaderTimeout() time.Duration
	Timeout() time.Duration
	CORSAllowedOrigins() []string
}
