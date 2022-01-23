package kernel

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	_ KeyValueDB = &KeyValueDBT{}
	_ Iterator   = &IteratorT{}
)

type KeyValueDBT struct {
	ptr *leveldb.DB
}

func NewDB(path string) KeyValueDB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil
	}
	return &KeyValueDBT{ptr: db}
}

func (db *KeyValueDBT) Iter(prefix []byte) Iterator {
	return &IteratorT{
		ptr: db.ptr.NewIterator(util.BytesPrefix(prefix), nil),
	}
}

func (db *KeyValueDBT) Set(key []byte, value []byte) {
	err := db.ptr.Put(key, value, nil)
	if err != nil {
		panic(err)
	}
}

func (db *KeyValueDBT) Get(key []byte) []byte {
	data, err := db.ptr.Get(key, nil)
	if err != nil {
		return nil
	}
	return data
}

func (db *KeyValueDBT) Del(key []byte) {
	err := db.ptr.Delete(key, nil)
	if err != nil {
		panic(err)
	}
}

func (db *KeyValueDBT) Close() {
	db.ptr.Close()
}

type IteratorT struct {
	ptr iterator.Iterator
}

func (iter *IteratorT) Next() bool {
	return iter.ptr.Next()
}

func (iter *IteratorT) Key() []byte {
	return iter.ptr.Key()
}

func (iter *IteratorT) Value() []byte {
	return iter.ptr.Value()
}

func (iter *IteratorT) Close() {
	iter.ptr.Release()
}
