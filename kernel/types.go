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
	Find(Hash) Object
	Append(Object) error
	Length() BigInt
	Range(BigInt, BigInt) Objects
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
	PrevHash() Hash
	Accept(PrivKey) error

	Editor
	Wrapper
	Signifier
}

type Laziness interface {
	Interval(PubKey) BigInt
	SelectLazy([]PubKey) (PubKey, BigInt)
}

type Chain interface {
	LastHash() Hash
	NonceIsValid(Block, Transaction) bool

	Laziness
	Editor
	Wrapper
	Verifier
}
