package kernel

import (
	"bytes"
	"path/filepath"
	"sort"
	"sync"

	"github.com/number571/go-peer/encoding"
)

var (
	_ Chain = &ChainT{}
)

type ChainT struct {
	mtx     sync.Mutex
	path    string
	blocks  KeyValueDB
	txs     KeyValueDB
	mempool Mempool
}

func NewChain(path string, genesis Block) Chain {
	var (
		blocksPath  = filepath.Join(path, BlocksPath)
		txsPath     = filepath.Join(path, TXsPath)
		mempoolPath = filepath.Join(path, MempoolPath)
	)

	if !genesis.IsValid() {
		return nil
	}

	blocks := NewDB(blocksPath)
	if blocks == nil {
		return nil
	}

	txs := NewDB(txsPath)
	if txs == nil {
		return nil
	}

	mempool := NewDB(mempoolPath)
	if mempool == nil {
		return nil
	}

	chain := &ChainT{
		path:   path,
		blocks: blocks,
		txs:    txs,
		mempool: &MempoolT{
			ptr: mempool,
		},
	}

	chain.setHeight(0)
	chain.setBlock(genesis)
	mempool.Set(GetKeyMempoolHeight(), encoding.Uint64ToBytes(0))

	return chain
}

func LoadChain(path string) Chain {
	var (
		blocksPath  = filepath.Join(path, BlocksPath)
		txsPath     = filepath.Join(path, TXsPath)
		mempoolPath = filepath.Join(path, MempoolPath)
	)

	blocks := NewDB(blocksPath)
	if blocks == nil {
		return nil
	}

	txs := NewDB(txsPath)
	if txs == nil {
		return nil
	}

	mempool := NewDB(mempoolPath)
	if mempool == nil {
		return nil
	}

	return &ChainT{
		path:   path,
		blocks: blocks,
		txs:    txs,
		mempool: &MempoolT{
			ptr: mempool,
		},
	}
}

func (chain *ChainT) Close() {
	chain.blocks.Close()
	chain.txs.Close()

	mempool := chain.mempool.(*MempoolT)
	mempool.ptr.Close()
}

func (chain *ChainT) Mempool() Mempool {
	return chain.mempool
}

func (chain *ChainT) Accept(block Block) bool {
	chain.mtx.Lock()
	defer chain.mtx.Unlock()

	if block == nil {
		return false
	}

	if !block.IsValid() {
		return false
	}

	lastBlock := chain.Block(chain.Height())
	if !bytes.Equal(lastBlock.Hash(), block.PrevHash()) {
		return false
	}

	for _, tx := range block.Transactions() {
		if chain.TX(tx.Hash()) != nil {
			return false
		}
	}

	mempool := chain.Mempool()
	for _, tx := range block.Transactions() {
		mempool.Delete(tx.Hash())
	}

	chain.setHeight(chain.Height() + 1)
	chain.setBlock(block)

	return true
}

func (chain *ChainT) Merge(height Height, txs []Transaction) bool {
	chain.mtx.Lock()
	defer chain.mtx.Unlock()

	var (
		lastBlock = chain.Block(height)
		resultTXs []Transaction
	)

	if chain.Height() != height {
		return false
	}

	resultTXs = append(resultTXs, lastBlock.Transactions()...)

	for _, tx := range txs {
		if !tx.IsValid() {
			return false
		}

		if chain.TX(tx.Hash()) != nil {
			continue
		}

		resultTXs = append(resultTXs, tx)
	}

	if len(resultTXs) == TXsSize {
		return false
	}

	// select x transactions from X by algorithm
	sort.SliceStable(resultTXs, func(i, j int) bool {
		return bytes.Compare(resultTXs[i].Hash(), resultTXs[j].Hash()) < 0
	})

	appendTXs := resultTXs[:TXsSize]
	deleteTXs := resultTXs[TXsSize:]

	chain.updateBlock(height, NewBlock(lastBlock.PrevHash(), appendTXs), deleteTXs)
	return true
}

func (chain *ChainT) Height() Height {
	return chain.getHeight()
}

func (chain *ChainT) TX(hash Hash) Transaction {
	return chain.getTX(hash)
}

func (chain *ChainT) Block(height Height) Block {
	return chain.getBlock(height)
}

// Height

func (chain *ChainT) getHeight() Height {
	data := chain.blocks.Get(GetKeyHeight())
	if data == nil {
		panic("chain: height undefined")
	}
	return Height(encoding.BytesToUint64(data))
}

func (chain *ChainT) setHeight(height Height) {
	chain.blocks.Set(GetKeyHeight(), encoding.Uint64ToBytes(uint64(height)))
}

// TX

func (chain *ChainT) getTX(hash Hash) Transaction {
	data := chain.txs.Get(GetKeyTX(hash))
	return LoadTransaction(data)
}

func (chain *ChainT) setTX(tx Transaction) {
	chain.txs.Set(GetKeyTX(tx.Hash()), tx.Bytes())
}

func (chain *ChainT) delTX(tx Transaction) {
	chain.txs.Del(GetKeyTX(tx.Hash()))
}

// Block

func (chain *ChainT) getBlock(height Height) Block {
	data := chain.blocks.Get(GetKeyBlock(height))
	return LoadBlock(data)
}

func (chain *ChainT) setBlock(block Block) {
	chain.blocks.Set(GetKeyBlock(chain.Height()), block.Bytes())

	for _, tx := range block.Transactions() {
		chain.setTX(tx)
	}
}

func (chain *ChainT) updateBlock(height Height, block Block, delTXs []Transaction) {
	chain.blocks.Set(GetKeyBlock(height), block.Bytes())

	for _, tx := range block.Transactions() {
		chain.setTX(tx)
		go chain.Mempool().Delete(tx.Hash())
	}

	for _, tx := range delTXs {
		chain.delTX(tx)
		go chain.Mempool().Push(tx)
	}
}
