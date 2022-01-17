package kernel

import (
	"github.com/number571/go-peer/crypto"
)

type Height uint64
type Hash []byte
type Sign []byte

type PrivKey crypto.PrivKey
type PubKey crypto.PubKey

type Wrapper interface {
	Bytes() []byte
	String() string
}

type Hasher interface {
	Hash() Hash
	IsValid() bool
}

type Signifier interface {
	Sign() Sign
	Validator() PubKey

	Hasher
}

type KeyValueDB interface {
	Set([]byte, []byte)
	Get([]byte) []byte
	Del([]byte)
	Close()
}

type Mempool interface {
	Height() Height
	TX(Hash) Transaction
	Clear(Hash)

	Push(Transaction)
	Pop() []Transaction
}

type Chain interface {
	Accept(Block) bool
	Merge(Height, []Transaction) bool
	Height() Height

	TX(Hash) Transaction
	Block(Height) Block
	Mempool() Mempool

	Close()
}

type Block interface {
	PrevHash() Hash
	Transactions() []Transaction

	Wrapper
	Hasher
}

type Transaction interface {
	PayLoad() []byte

	Wrapper
	Signifier
}
