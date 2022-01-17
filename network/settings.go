package network

const (
	BuffSize  = (1 << 10) // 1KiB
	PackSize  = (1 << 20) // 1MiB
	ConnSize  = 256       // max num connections
	RetrySize = 5         // num retry send
	TimeLimit = 5         // seconds
)
