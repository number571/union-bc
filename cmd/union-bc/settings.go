package main

import (
	"log"

	"github.com/number571/union-bc/kernel"
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
	IntervalTime = 1 // seconds
)

type Logger struct {
	reset   string
	message string
}

func Log() *Logger {
	return &Logger{
		reset:   "\033[0m",
		message: "[%c] %-12sheight=%020d hash=%64X mempool=%05d txs=%d",
	}
}

func (lg *Logger) Warning(name string, height kernel.Height, hash []byte, mempool kernel.Height, txs uint32) {
	colorYellow := "\033[33m"
	log.Printf(colorYellow+lg.message+lg.reset,
		'W', name, height, hash, mempool, txs)
}

func (lg *Logger) Error(name string, height kernel.Height, hash []byte, mempool kernel.Height, txs uint32) {
	colorRed := "\033[31m"
	log.Printf(colorRed+lg.message+lg.reset,
		'E', name, height, hash, mempool, txs)
}

func (lg *Logger) Info(name string, height kernel.Height, hash []byte, mempool kernel.Height, txs uint32) {
	log.Printf(lg.message+lg.reset,
		'I', name, height, hash, mempool, txs)
}
