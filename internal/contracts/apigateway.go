package contracts

import (
	"time"
)

type APIGateway interface {
	ListenAddr() string
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	Timeout() time.Duration
}
