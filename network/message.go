package network

import (
	"bytes"
	"encoding/json"

	"github.com/number571/go-peer/crypto"
)

var (
	_ Message = &MessageT{}
)

type MessageT struct {
	HeadT    MsgType `json:"head"`
	BodyT    []byte  `json:"body"`
	NonceT   []byte  `json:"nonce"`
	NetworkT string  `json:"network"`
}

// Create message with title and data.
func NewMessage(head MsgType, body []byte) Message {
	return &MessageT{
		HeadT:    head,
		BodyT:    body,
		NonceT:   crypto.RandBytes(16),
		NetworkT: NetworkName,
	}
}

func (msg *MessageT) Head() MsgType {
	return msg.HeadT
}

func (msg *MessageT) Body() []byte {
	return msg.BodyT
}

func (msg *MessageT) Nonce() []byte {
	return msg.NonceT
}

func (msg *MessageT) Network() string {
	return msg.NetworkT
}

func (msg *MessageT) Hash() string {
	return crypto.NewSHA256(msg.Bytes()).String()
}

// Serialize with JSON format.
func (msg *MessageT) Bytes() []byte {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil
	}

	pack := PackageT(jsonData)
	return bytes.Join(
		[][]byte{
			pack.SizeToBytes(),
			pack.Bytes(),
		},
		[]byte{},
	)
}
