package protomake

import (
	"context"
	"sync"
)

// TxHandler represents event handler at transmitter side.
type TxHandler interface {
	// Detect returns true when this handler needs to handle the event.
	Detect(evt interface{}) bool
	// Handle handles event from inbound channel (evtChan) and emits message to outband channel (msgChan).
	Handle(ctx context.Context, evtChan <-chan interface{}) (msgChan <-chan interface{})
}

// NewTxDispatcher creates a new TxDispatcher instance.
func NewTxDispatcher(msgChan chan<- interface{}) *TxDispatcher {
	return &TxDispatcher{
		msgChan: msgChan,
	}
}

// TxDispatcher represents event dispatcher at transmitter side.
type TxDispatcher struct {
	evtChan  <-chan interface{}
	msgChan  chan<- interface{}
	handlers []TxHandler
}

// Register registers TxHandler to TxDispatcher.
func (d *TxDispatcher) Register(h TxHandler) {
	d.handlers = append(d.handlers, h)
}

func (d *TxDispatcher) run(ctx context.Context) {

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

		for evt := range d.evtChan {
			select {
			case <-ctx.Done():
				return
			default:
			}

			for i, h := range d.handlers {
				if h.Detect(evt) {
					toChans[i] <- evt
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

			for msg := range fromChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				d.msgChan <- msg
			}
		}()

	}

	wg.Wait()
}
