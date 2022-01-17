package main

import "github.com/number571/unionbc/network"

const (
	MsgGetHeight network.MsgType = iota + 1
	MsgGetBlock
	MsgSetBlock
	MsgGetTX
	MsgSetTX
)

const (
	MaskBit = network.MsgType(1 << 31)
)