package main

import "time"

type PingMessage struct {
	TxTime time.Time
}

type PongMessage struct {
	TxTime time.Time
	RxTime time.Time
}

type PingGotEvent struct {
	RxTime time.Time
	TxTime time.Time
}
