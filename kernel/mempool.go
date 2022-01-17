package kernel

import (
	"sync"

	"github.com/number571/go-peer/encoding"
	"github.com/syndtr/goleveldb/leveldb/util"
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
		panic("value undefined")
	}
	return Height(encoding.BytesToUint64(data))
}

func (mempool *MempoolT) TX(hash Hash) Transaction {
	data := mempool.ptr.Get(GetKeyMempoolTX(hash))
	return LoadTransaction(data)
}

func (mempool *MempoolT) Clear(hash Hash) {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	mempool.deleteTX(hash)
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

func (mempool *MempoolT) Pop() Transaction {
	mempool.mtx.Lock()
	defer mempool.mtx.Unlock()

	var (
		db      = (mempool.ptr).(*KeyValueDBT)
		iter    = db.ptr.NewIterator(util.BytesPrefix([]byte(KeyMempoolPrefixTX)), nil)
		txBytes []byte
	)

	for iter.Next() {
		txBytes = iter.Value()
		break
	}
	iter.Release()

	tx := LoadTransaction(txBytes)
	if tx == nil {
		return nil
	}

	mempool.deleteTX(tx.Hash())
	return tx
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
