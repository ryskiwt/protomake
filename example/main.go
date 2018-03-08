package main

import (
	"context"
	"fmt"
	"time"

	"pgithub.com/ryskiwt/protomake"
)

func main() {
	recvChan := make(chan interface{}, 1024)
	sendChan := make(chan interface{}, 1024)

	var r rx
	rd := protomake.NewRxDispatcher(recvChan)
	rd.Register(&r)

	var t tx
	td := protomake.NewTxDispatcher(sendChan)
	td.Register(&t)

	var i ic
	id := protomake.NewInterceptDispatcher()
	id.Register(&i)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		protomake.Serve(ctx, rd, id, td)
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			msg := PingMessage{
				TxTime: <-ticker.C,
			}

			fmt.Printf("RECV: %#v\n", msg)
			recvChan <- msg
		}
	}()

	for msg := range sendChan {
		fmt.Printf("SEND: %#v\n\n", msg)
	}
}
