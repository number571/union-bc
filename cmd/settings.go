package main

import (
	"github.com/number571/union-bc/network"
)

const (
	MsgGetTime   = 1
	MsgGetHeight = 2
	MsgGetBlock  = 3
	MsgSetBlock  = 4
	MsgGetTX     = 5
	MsgSetTX     = 6
)

const (
	MaskBit      = network.MsgType(1 << 31)
	IntervalTime = 5 // seconds
	ClientsNum   = 3
	TXsInSecond  = 3
)
