package stream

import "time"

type WithProxyContract interface {
	Method() string
	Path() string
	Host() string
	ServiceName() string
	HeadersToJson() string
	Duration() time.Duration
}
