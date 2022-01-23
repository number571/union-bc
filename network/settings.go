package network

const (
	MappSize  = 2048      // hashes
	ConnSize  = 512       // max num connections
	RetrySize = 8         // num retry send
	PackSize  = (2 << 20) // 2MiB
)

const (
	IsNode   byte = 1
	IsClient byte = 2
)
