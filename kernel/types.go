package kernel

import (
	"github.com/number571/gopeer/crypto"
)

type Hash []byte
type Sign []byte

type PrivKey crypto.PrivKey
type PubKey crypto.PubKey

type Object interface{}

type BigInt interface {
	Inc() BigInt
	Bytes() []byte
	String() string
	Cmp(BigInt) int
}

type Verifier interface {
	IsValid() bool
}

type Signifier interface {
	Hash() Hash
	Sign() Sign
	Validator() PubKey
	Verifier
}

type Editor interface {
	Append(Object) error
	Find(Hash) Object
}

type Wrapper interface {
	Wrap() []byte
}

type Transaction interface {
	Nonce() BigInt
	Data() []byte

	Wrapper
	Signifier
}

type Block interface {
	TXs() []Transaction
	PrevHash() Hash
	Accept(PrivKey) error

	Editor
	Wrapper
	Signifier
}

type Chain interface {
	Blocks() []Block
	LastHash() Hash
	Interval(PubKey) BigInt
	NonceIsValid(Block, Transaction) bool

	Editor
	Wrapper
	Verifier
}
