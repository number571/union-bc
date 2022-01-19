package network

const (
	MappSize  = 1024      // hashes
	BuffSize  = (1 << 10) // 1KiB
	PackSize  = (1 << 20) // 1MiB
	ConnSize  = 256       // max num connections
	RetrySize = 32        // num retry send
	TimeLimit = 5         // seconds
)

const (
	IsNode   byte = 1
	IsClient byte = 2
)
