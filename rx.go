package protomake

import (
	"context"
	"sync"
)

// RxHandler represents message handler at receiver side.
type RxHandler interface {
	// Detect returns true when this handler needs to handle the message
	Detect(msg interface{}) bool
	// Handle handles message from inbound channel (msgChan) and emits event to outbound channel (evtChan).
	Handle(ctx context.Context, msgChan <-chan interface{}) (evtChan <-chan interface{})
}

// NewRxDispatcher creates a new RxDispatcher instance.
func NewRxDispatcher(msgChan <-chan interface{}) *RxDispatcher {
	return &RxDispatcher{
		msgChan: msgChan,
	}
}

// RxDispatcher represents message dispatcher at receiver side.
type RxDispatcher struct {
	msgChan  <-chan interface{}
	evtChan  chan<- interface{}
	handlers []RxHandler
}

// Register registers handlers.
func (d *RxDispatcher) Register(h RxHandler) {
	d.handlers = append(d.handlers, h)
}

func (d *RxDispatcher) run(ctx context.Context) {

	//
	// make chans
	//

	toChans := make([]chan interface{}, 0, len(d.handlers))
	fromChans := make([]<-chan interface{}, 0, len(d.handlers))
	for _, h := range d.handlers {
		toChan := make(chan interface{}, len(d.evtChan))
		toChans = append(toChans, toChan)
		fromChan := h.Handle(ctx, toChan)
		fromChans = append(fromChans, fromChan)
	}

	//
	// fan-out
	//

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for msg := range d.msgChan {
			select {
			case <-ctx.Done():
				return
			default:
			}

			for i, h := range d.handlers {
				if h.Detect(msg) {
					toChans[i] <- msg
				}
			}
		}
	}()

	//
	// fan-in
	//

	for i := range fromChans {
		fromChan := fromChans[i]

		wg.Add(1)
		go func() {
			defer wg.Done()

			for evt := range fromChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				d.evtChan <- evt
			}
		}()

	}

	wg.Wait()
}
