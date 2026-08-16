package stream

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/serge1997/apigateway/internal/database"
	"gorm.io/gorm/clause"
)

func consumer() {
	go func() {
		var batchedStream []*RequestStream = []*RequestStream{}
		var doneStream = make(chan struct{}, 1)
		persistFn := func() {
			if len(batchedStream) >= 1 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				database.Db().WithContext(ctx).Clauses(clause.OnConflict{
					DoNothing: true,
				}).Create(&batchedStream)
			}
			batchedStream = batchedStream[:0]
			select {
			case doneStream <- struct{}{}:
			default:
			}
		}
	loop:
		for {
			timeout := time.After(restartTimeout)
			select {
			case toPerisistStream := <-dataStream:
				batchedStream = append(batchedStream, toPerisistStream)
				fmt.Printf("%+v\n", batchedStream[0])
				if len(batchedStream) >= batchSize {
					persistFn()
				}
			case <-doneStream:
				log.Println("[worker - healthy] worker executed")
				time.Sleep(time.Millisecond * 1500)
			case <-timeout:
				persistFn()
				fmt.Println("restarting")
				continue loop
			}
		}
	}()
}
