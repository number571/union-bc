package kernel

import (
	"github.com/number571/gopeer/crypto"
)

type Hash []byte
type Sign []byte

type PrivKey crypto.PrivKey
type PubKey crypto.PubKey

type Object interface{}

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
	Range(BigInt, BigInt) Object
}

type Wrapper interface {
	Wrap() []byte
}

type BigInt interface {
	Inc() BigInt
	Dec() BigInt

	Sub(BigInt) BigInt
	Cmp(BigInt) int

	Bytes() []byte
	String() string
	Uint64() uint64
}

type Transaction interface {
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
	LazyInterval(PubKey) BigInt
	SelectLazy([]PubKey) PubKey

	RollBack(BigInt)
	Close()

	Editor
	Verifier
}
