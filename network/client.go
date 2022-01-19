package network

import (
	"net"
)

type ClientT struct {
	conn Conn
}

func NewClient(address string) Client {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}

	conn.Write([]byte{IsClient})
	return &ClientT{conn}
}

func (client *ClientT) Request(msg Message) Message {
	client.conn.Write(msg.Bytes())
	return ReadMessage(client.conn)
}

func (client *ClientT) Close() {
	client.conn.Close()
}
