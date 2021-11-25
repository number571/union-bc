package kernel

import (
	"github.com/number571/gopeer/crypto"
)

type Hash []byte
type Sign []byte

type PrivKey crypto.PrivKey
type PubKey crypto.PubKey

type Object interface{}
type Objects interface{}

type BigInt interface {
	Inc() BigInt
	Bytes() []byte
	String() string
	Cmp(BigInt) int
	Uint64() uint64
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
	LastHash() Hash
	Find(Hash) Object
	Append(Object) error
	Length() BigInt
	Range(BigInt, BigInt) Objects
}

type Wrapper interface {
	Wrap() []byte
}

type Laziness interface {
	LazyInterval(PubKey) BigInt
	SelectLazy([]PubKey) PubKey
}

type Transaction interface {
	Nonce() BigInt
	Data() []byte

	Wrapper
	Signifier
}

type Block interface {
	Accept(PrivKey) error

	Editor
	Wrapper
	Signifier
}

type Chain interface {
	NonceIsValid(Block, Transaction) bool
	Laziness

	Editor
	Wrapper
	Verifier
}
