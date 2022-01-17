package kernel

import (
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	_ KeyValueDB = &KeyValueDBT{}
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
