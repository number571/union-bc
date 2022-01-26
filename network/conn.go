package network

import (
	"bytes"
	"encoding/json"
	"time"
)

func ReadMessage(conn Conn) Message {
	ch := make(chan Message)
	go readMessage(conn, ch)

	select {
	case rmsg := <-ch:
		return rmsg
	case <-time.After(TimeSize * time.Second):
		return nil
	}
}

func readMessage(conn Conn, ch chan Message) {
	const (
		SizeUint64 = 8 // bytes
	)

	var (
		pack   []byte
		size   = uint64(0)
		msg    = new(MessageT)
		buflen = make([]byte, SizeUint64)
	)

	length, err := conn.Read(buflen)
	if err != nil {
		// fmt.Println(111)
		ch <- nil
		return
	}

	if length != SizeUint64 {
		// fmt.Println(222)
		ch <- nil
		return
	}

	mustLen := PackageT(buflen).BytesToSize()
	if mustLen > PackSize {
		// fmt.Println(333)
		ch <- nil
		return
	}

	for {
		buffer := make([]byte, mustLen-size)

		length, err = conn.Read(buffer)
		if err != nil {
			// fmt.Println(444)
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
		ch <- nil
		return
	}

	if msg.Network() != NetworkName {
		ch <- nil
		return
	}

	ch <- msg
}
