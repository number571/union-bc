package main

import (
	"log"

	"github.com/number571/union-bc/kernel"
)

type Logger struct {
	reset   string
	message string
}

func Log() *Logger {
	return &Logger{
		reset:   "\033[0m",
		message: "[%c] %-10sheight=%012d hash=...%032X mempool=%06d txs=%d conn=%d",
	}
}

func (lg *Logger) Warning(name string, height kernel.Height, hash []byte, mempool kernel.Height, txs int, conns int) {
	colorYellow := "\033[33m"
	log.Printf(colorYellow+lg.message+lg.reset,
		'W', name, height, hash[16:], mempool, txs, conns)
}

func (lg *Logger) Error(name string, height kernel.Height, mempool kernel.Height, txs int, conns int) {
	colorRed := "\033[31m"
	log.Printf(colorRed+lg.message+lg.reset,
		'E', name, height, []byte{0}, mempool, txs, conns)
}

func (lg *Logger) Info(name string, height kernel.Height, hash []byte, mempool kernel.Height, txs int, conns int) {
	log.Printf(lg.message+lg.reset,
		'I', name, height, hash[16:], mempool, txs, conns)
}
