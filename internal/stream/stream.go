package stream

import (
	"time"
)

var dataStream = make(chan *RequestStream, 100)
var batchSize = 2
var restartTimeout = time.Second * 10
var (
	HeaderContentType string = "Content-Type"
	HeaderUserAgent   string = "User-Agent"
)

func init() {
	consumer()
}
func Produce(data *RequestStream) {
	select {
	case dataStream <- data:
	default:
	}
}
