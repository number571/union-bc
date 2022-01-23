package network

import (
	"net"
	"sync"
)

type Conn net.Conn
type MsgType uint32

type HandleFunc func(Node, Conn, Message)

type Message interface {
	Head() MsgType
	Body() []byte
	Nonce() []byte

	Hash() string
	Bytes() []byte
}

type Package interface {
	Size() uint64
	Bytes() []byte

	SizeToBytes() []byte
	BytesToSize() uint64
}

type Client interface {
	Request(Message) Message
	Close()
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
