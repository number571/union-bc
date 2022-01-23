package network

import (
	"bytes"
	"encoding/json"
)

func ReadMessage(conn Conn) Message {
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
		return nil
	}

	if length != SizeUint64 {
		// fmt.Println(222)
		return nil
	}

	mustLen := PackageT(buflen).BytesToSize()
	if mustLen > PackSize {
		// fmt.Println(333)
		return nil
	}

	for {
		buffer := make([]byte, mustLen-size)

		length, err = conn.Read(buffer)
		if err != nil {
			// fmt.Println(444)
			return nil
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
		// fmt.Println(555)
		return nil
	}

	return msg
}
