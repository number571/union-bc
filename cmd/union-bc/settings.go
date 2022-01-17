package main

import "github.com/number571/union-bc/network"

const (
	MsgGetTime network.MsgType = iota + 1
	MsgGetHeight
	MsgGetBlock
	MsgSetBlock
	MsgGetTX
	MsgSetTX
)

const (
	MaskBit = network.MsgType(1 << 31)
)
