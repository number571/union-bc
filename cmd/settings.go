package main

const (
	MsgGetTime   = 0x01
	MsgGetHeight = 0x02
	MsgGetBlock  = 0x03
	MsgSetBlock  = 0x04
	MsgGetTX     = 0x05
	MsgSetTX     = 0x06
)

const (
	MaskBit      = (1 << 31)
	IntervalTime = 5 // seconds
	ClientsNum   = 3
	TXsInSecond  = 3
)
