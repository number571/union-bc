package kernel

import (
	"sync"

	"github.com/number571/go-peer/encoding"
)

var (
	_ Mempool = &MempoolT{}
)

type MempoolT struct {
	mtx sync.Mutex
	ptr KeyValueDB
}

func (mempool *MempoolT) Height() Height {
	data := mempool.ptr.Get(GetKeyMempoolHeight())
	if data == nil {
		panic("mempool: height undefined")
	}
	return Height(encoding.BytesToUint64(data))
}

func (mempool *MempoolT) TX(hash Hash) Transaction {
	data := mempool.ptr.Get(GetKeyMempoolTX(hash))
	return LoadTransaction(data)
}

func (mempool *MempoolT) Delete(hash Hash) {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	mempool.deleteTX(hash)
}

func (mempool *MempoolT) Clear() {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	iter := mempool.ptr.Iter([]byte(KeyMempoolPrefixTX))
	defer iter.Close()

	for iter.Next() {
		txBytes := iter.Value()

		tx := LoadTransaction(txBytes)
		if tx == nil {
			panic("mempool: tx is nil")
		}

		mempool.deleteTX(tx.Hash())
	}
}

func (mempool *MempoolT) Push(tx Transaction) {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	var (
		hash      = tx.Hash()
		newHeight = uint64(mempool.Height() + 1)
	)

	if newHeight > MempoolSize {
		return
	}

	if mempool.TX(hash) != nil {
		return
	}

	mempool.ptr.Set(GetKeyMempoolHeight(), encoding.Uint64ToBytes(newHeight))
	mempool.ptr.Set(GetKeyMempoolTX(hash), tx.Bytes())
}

func (mempool *MempoolT) Pop() []Transaction {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	if mempool.Height() < TXsSize {
		return nil
	}

	var (
		txs   []Transaction
		count uint
	)

	iter := mempool.ptr.Iter([]byte(KeyMempoolPrefixTX))
	defer iter.Close()

	for count = 0; iter.Next() && count < TXsSize; count++ {
		txBytes := iter.Value()

		tx := LoadTransaction(txBytes)
		if tx == nil {
			return nil
		}

		txs = append(txs, tx)
	}

	if count != TXsSize {
		return nil
	}

	for _, tx := range txs {
		mempool.deleteTX(tx.Hash())
	}

	return txs
}

func (mempool *MempoolT) deleteTX(hash Hash) {
	var (
		newHeight = uint64(mempool.Height() - 1)
	)

	if mempool.TX(hash) == nil {
		return
	}

	mempool.ptr.Set(GetKeyMempoolHeight(), encoding.Uint64ToBytes(newHeight))
	mempool.ptr.Del(GetKeyMempoolTX(hash))
}
