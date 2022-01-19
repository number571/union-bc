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
		size   = uint(0)
		msg    = new(MessageT)
		buflen = make([]byte, SizeUint64)
		buffer = make([]byte, PackSize)
	)

	length, err := conn.Read(buflen)
	if err != nil {
		return nil
	}
	if length != SizeUint64 {
		return nil
	}

	mustLen := PackageT(buflen).BytesToSize()
	if mustLen > PackSize {
		return nil
	}

	for {
		length, err = conn.Read(buffer)
		if err != nil {
			return nil
		}

		size += uint(length)
		if size > mustLen {
			return nil
		}

		pack = bytes.Join(
			[][]byte{
				pack,
				buffer[:length],
			},
			[]byte{},
		)

		if size == mustLen {
			break
		}
	}

	err = json.Unmarshal(pack, msg)
	if err != nil {
		return nil
	}

	return msg
}
