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

type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
	Close()
}

type KeyValueDB interface {
	Iter([]byte) Iterator
	Set([]byte, []byte)
	Get([]byte) []byte
	Del([]byte)
	Close()
}

type Mempool interface {
	Height() Height
	TX(Hash) Transaction

	Push(Transaction)
	Pop() []Transaction

	Delete(Hash)
	Clear()
}

type Chain interface {
	Accept(Block) bool
	Merge(Height, []Transaction) bool
	Rollback(uint64) bool

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
