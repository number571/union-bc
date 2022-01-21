package kernel

const (
	KeySize     = 1024 // num bits
	MempoolSize = 512  // max num txs in mempool
	TXsSize     = 16   // num txs in block

	BlocksPath  = "blocks.db"
	TXsPath     = "txs.db"
	MempoolPath = "mempool.db"

	KeyHeight = "chain.blocks.height"
	KeyBlock  = "chain.blocks.block[%d]"
	KeyTX     = "chain.txs.tx[%X]"

	KeyMempoolHeight   = "chain.mempool.height"
	KeyMempoolTX       = "chain.mempool.tx[%X]"
	KeyMempoolPrefixTX = "chain.mempool.tx["
)
