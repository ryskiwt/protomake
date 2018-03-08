package main

import (
	"context"
	"time"
)

type rx struct{}

func (x *rx) Detect(msg interface{}) bool {
	_, ok := msg.(PingMessage)
	return ok
}

func (x *rx) Handle(ctx context.Context, msgChan <-chan interface{}) <-chan interface{} {
	evtChan := make(chan interface{}, len(msgChan))

	go func() {
		for msg := range msgChan {
			select {
			case <-ctx.Done():
			default:
			}

			m := msg.(PingMessage)
			evtChan <- PingGotEvent{
				RxTime: time.Now(),
				TxTime: m.TxTime,
			}
		}
	}()

	return evtChan
}
