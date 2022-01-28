package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/number571/go-peer/crypto"
)

var (
	_ Conn = &ConnT{}
)

type ConnT struct {
	nonce string
	ptr   net.Conn
}

func NewConn(address string) Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}

	conn.Write([]byte{IsClient})
	return &ConnT{crypto.RandString(16), conn}
}

func (conn *ConnT) Request(msg Message) Message {
	conn.Write(msg)
	return conn.Read()
}

func (conn *ConnT) Close() error {
	return conn.ptr.Close()
}

func (conn *ConnT) Write(msg Message) {
	conn.ptr.Write(msg.Bytes())
}

func (conn *ConnT) Read() Message {
	ch := make(chan Message)
	go readMessage(conn, ch)

	select {
	case rmsg := <-ch:
		return rmsg
	case <-time.After(TimeSize * time.Second):
		fmt.Println(777)
		return nil
	}
}

func readMessage(conn *ConnT, ch chan Message) {
	const (
		SizeUint64 = 8 // bytes
	)

	var (
		pack   []byte
		size   = uint64(0)
		msg    = new(MessageT)
		buflen = make([]byte, SizeUint64)
	)

	length, err := conn.ptr.Read(buflen)
	if err != nil {
		// fmt.Println(111)
		ch <- nil
		return
	}

	if length != SizeUint64 {
		fmt.Println(222)
		ch <- nil
		return
	}

	mustLen := PackageT(buflen).BytesToSize()
	if mustLen > PackSize {
		fmt.Println(333)
		ch <- nil
		return
	}

	for {
		buffer := make([]byte, mustLen-size)

		length, err = conn.ptr.Read(buffer)
		if err != nil {
			fmt.Println(444)
			ch <- nil
			return
		}

		pack = bytes.Join(
			[][]byte{
				pack,
				buffer[:length],
			},
			[]byte{},
		)

		size += uint64(length)
		if size == mustLen {
			break
		}
	}

	err = json.Unmarshal(pack, msg)
	if err != nil {
		fmt.Println(555)
		ch <- nil
		return
	}

	if msg.Network() != NetworkName {
		fmt.Println(666)
		ch <- nil
		return
	}

	ch <- msg
}
