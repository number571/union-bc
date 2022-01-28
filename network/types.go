package network

import (
	"sync"
)

type MsgType uint32
type HandleFunc func(Node, Conn, Message)

type Message interface {
	Head() MsgType
	Body() []byte

	Nonce() []byte
	Network() string

	Hash() string
	Bytes() []byte
}

type Package interface {
	Size() uint64
	Bytes() []byte

	SizeToBytes() []byte
	BytesToSize() uint64
}

type Conn interface {
	Request(Message) Message
	Close() error

	Write(Message)
	Read() Message
}

type Node interface {
	Mutex() *sync.Mutex

	Broadcast(Message)
	Listen(string) error
	Handle(MsgType, HandleFunc) Node

	Connect(string) Conn
	Disconnect(Conn)
	Connections() []Conn
}
