package main

import "context"

type tx struct{}

func (x *tx) Detect(evt interface{}) bool {
	_, ok := evt.(PingGotEvent)
	return ok
}

func (x *tx) Handle(ctx context.Context, evtChan <-chan interface{}) <-chan interface{} {
	msgChan := make(chan interface{}, len(evtChan))

	go func() {
		for evt := range evtChan {
			select {
			case <-ctx.Done():
			default:
			}

			e := evt.(PingGotEvent)
			msgChan <- PongMessage{
				TxTime: e.TxTime,
				RxTime: e.RxTime,
			}
		}
	}()

	return msgChan
}
