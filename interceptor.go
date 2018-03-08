package protomake

import (
	"context"
	"sync"
)

// InterceptHandler represents entercept handler.
type InterceptHandler interface {
	// Detect returns true when this handler needs to handle the event.
	Detect(evt interface{}) bool
	// Handle handles event from inbound channel (inEvtChan) and emits event to outbound channel (outevtChan).
	Handle(ctx context.Context, inEvtChan <-chan interface{}) (outEvtChan <-chan interface{})
}

// NewInterceptDispatcher creates a new InterceptDispatcher instance.
func NewInterceptDispatcher() *InterceptDispatcher {
	return &InterceptDispatcher{}
}

// InterceptDispatcher represents event dispatcher.
type InterceptDispatcher struct {
	inEvtChan  <-chan interface{}
	outEvtChan chan<- interface{}
	handlers   []InterceptHandler
}

// Register registers handlers.
func (d *InterceptDispatcher) Register(h InterceptHandler) {
	d.handlers = append(d.handlers, h)
}

func (d *InterceptDispatcher) run(ctx context.Context) {
	//
	// make chans
	//

	toChans := make([]chan interface{}, 0, len(d.handlers))
	fromChans := make([]<-chan interface{}, 0, len(d.handlers))
	for _, h := range d.handlers {
		toChan := make(chan interface{}, len(d.inEvtChan))
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

		for msg := range d.inEvtChan {
			select {
			case <-ctx.Done():
				return
			default:
			}

			d.outEvtChan <- msg
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

				d.outEvtChan <- evt
			}
		}()

	}

	wg.Wait()
}
