package protomake

import (
	"context"
	"sync"
)

// Serve starts to serve.
func Serve(ctx context.Context, rd *RxDispatcher, id *InterceptDispatcher, td *TxDispatcher) {
	evtChan0 := make(chan interface{}, len(rd.msgChan))
	rd.evtChan = evtChan0
	id.inEvtChan = evtChan0

	evtChan1 := make(chan interface{}, len(td.msgChan))
	id.outEvtChan = evtChan1
	td.evtChan = evtChan1

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		rd.run(ctx)
	}()

	go func() {
		defer wg.Done()
		id.run(ctx)
	}()

	go func() {
		defer wg.Done()
		td.run(ctx)
	}()

	wg.Wait()
}
