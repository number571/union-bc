package network

import (
	"github.com/number571/go-peer/encoding"
)

var (
	_ Package = PackageT{}
)

type PackageT []byte

// Size of package in big endian bytes.
func (pack PackageT) Size() uint64 {
	return uint64(len(pack.Bytes()))
}

// Size of package in big endian bytes.
func (pack PackageT) SizeToBytes() []byte {
	return encoding.Uint64ToBytes(uint64(pack.Size()))
}

// From big endian bytes to uint size.
func (pack PackageT) BytesToSize() uint64 {
	return encoding.BytesToUint64(pack.Bytes())
}

// Bytes of package.
func (pack PackageT) Bytes() []byte {
	return []byte(pack)
}
