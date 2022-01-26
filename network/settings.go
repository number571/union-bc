package network

const (
	MappSize  = 2048      // hashes
	ConnSize  = 512       // max num connections
	RetrySize = 32        // num retry send
	TimeSize  = 5         // seconds
	PackSize  = (2 << 20) // 2MiB
)

const (
	NetworkName = "union-network"
)

const (
	IsNode   byte = 1
	IsClient byte = 2
)
