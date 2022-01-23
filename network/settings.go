package network

const (
	MappSize  = 2048      // hashes
	ConnSize  = 2048      // max num connections
	RetrySize = 32        // num retry send
	PackSize  = (8 << 20) // bytes
)

const (
	IsNode   byte = 1
	IsClient byte = 2
)
