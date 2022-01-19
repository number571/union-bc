package network

import (
	"net"
	"sync"
)

type Conn net.Conn
type Mutex *sync.Mutex
type MsgType uint32

type HandleFunc func(Node, Conn, []byte)

type Message interface {
	Head() MsgType
	Body() []byte
	Nonce() []byte

	Hash() string
	Bytes() []byte
}

type Package interface {
	Size() uint
	Bytes() []byte

	SizeToBytes() []byte
	BytesToSize() uint
}

type Client interface {
	Request(Message) Message
	Close()
}

type Node interface {
	Moniker() string
	Lock()
	Unlock()

	Broadcast(Message)
	Listen(string) error
	Handle(MsgType, HandleFunc) Node

	Connect(string) Conn
	Disconnect(Conn)
	Connections() []Conn
}
