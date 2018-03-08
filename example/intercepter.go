package main

import (
	"context"
	"fmt"
)

type ic struct{}

func (x *ic) Detect(evt interface{}) bool {
	_, ok := evt.(PingGotEvent)
	return ok
}

func (x *ic) Handle(ctx context.Context, inEvtChan <-chan interface{}) <-chan interface{} {
	outEvtChan := make(chan interface{}, len(inEvtChan))

	go func() {
		for evt := range inEvtChan {
			select {
			case <-ctx.Done():
			default:
			}

			fmt.Printf("INTERCEPT: %#v\n", evt)
		}
	}()

	return outEvtChan
}
