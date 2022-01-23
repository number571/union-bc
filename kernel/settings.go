package kernel

const (
	KeySize     = 1024 // num bits
	MempoolSize = 2000 // max num txs in mempool

	TXsSize     = 256  // num txs in block
	PayloadSize = 1024 // num bytes in tx.payload

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
