## protomake - Protocol Maker

A framework to build realtime protocol

## example

### Define receive handler
- receive handler **receives message** and **emits event**
- any cases listed below is OK
  - receive **some messages** and **emit one event**
  - receive **one message** and **emit some events**
  - receive **some messages** and **emit no events**
  - receive **no messages** and **emit some events**

```go
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
```

### Define Intercept Handler
- intercept handler **receives event** and **emits event**
- any cases listed below is OK
  - receive **some events** and **emit one event**
  - receive **one event** and **emit some events**
  - receive **some events** and **emit no events**
  - receive **no events** and **emit some events**


```go
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
```


### Define Transmit Handler
- transmit handler **receives event** and **emits message**
- any cases listed below is OK
  - receive **some events** and **emit one message**
  - receive **one event** and **emit some message**
  - receive **some events** and **emit no message**
  - receive **no events** and **emit some messages**

```go
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
```

### Start serve

- bind inbound / outbound channels to dispatcher
- register handlers to dispatcher
- then, start serving

```go

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
```
