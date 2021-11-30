package kernel

import "math/big"

var (
	_ BigInt = &BigIntT{}
)

type BigIntT big.Int

func NewInt(strnum string) BigInt {
	res, ok := big.NewInt(0).SetString(strnum, 10)
	if !ok {
		return nil
	}
	return (*BigIntT)(res)
}

func LoadInt(bytes []byte) BigInt {
	return (*BigIntT)(big.NewInt(0).SetBytes(bytes))
}

func ZeroInt() BigInt {
	return NewInt("0")
}

func (x *BigIntT) Inc() BigInt {
	return (*BigIntT)((*big.Int)(x).Add((*big.Int)(x), big.NewInt(1)))
}

func (x *BigIntT) Sub(y BigInt) BigInt {
	yn := (*big.Int)((y).(*BigIntT))
	return (*BigIntT)((*big.Int)(x).Sub((*big.Int)(x), yn))
}

func (x *BigIntT) Cmp(y BigInt) int {
	return (*big.Int)(x).Cmp((*big.Int)(y.(*BigIntT)))
}

func (x *BigIntT) Bytes() []byte {
	return (*big.Int)(x).Bytes()
}

func (x *BigIntT) String() string {
	return (*big.Int)(x).String()
}

func (x *BigIntT) Uint64() uint64 {
	return (*big.Int)(x).Uint64()
}
