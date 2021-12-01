package kernel

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	_ Chain = &ChainT{}
)

type ChainT struct {
	db *leveldb.DB
}


func NewChain(path string, genesis Block) Chain {
	if !genesis.IsValid() {
		return nil
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil
	}

	chain := &ChainT{
		db: db,
	}

	err = chain.setLength(NewInt("0"))
	if err != nil {
		return nil
	}

	err = chain.setBlock(genesis)
	if err != nil {
		return nil
	}

	return chain
}

func (chain *ChainT) Range(x, y BigInt) Object {
	blocks := []Block{}

	for x.Cmp(y) < 0 {
		block := chain.getBlockByID(x)
		if block == nil {
			return nil
		}
		blocks = append(blocks, block)
		x = x.Inc()
	}

	return blocks
}

func (chain *ChainT) Length() BigInt {
	return chain.getLength()
}

func (chain *ChainT) LastHash() Hash {
	return chain.getBlockByID(chain.Length()).Hash()
}

func (chain *ChainT) Append(obj Object) error {
	block := obj.(Block)
	if block == nil {
		return errors.New("block is null")
	}

	if !block.IsValid() {
		return errors.New("block is invalid")
	}

	if !bytes.Equal(block.LastHash(), chain.LastHash()) {
		return errors.New("relation is invalid")
	}

	return chain.setBlock(block)
}

func (chain *ChainT) Find(hash Hash) Object {
	return chain.getBlockByHash(hash)
}

// TODO: LevelDB -> Search blocks
func (chain *ChainT) IsValid() bool {
	// for _, block := range chain.blocks {
	// 	if !block.IsValid() {
	// 		return false
	// 	}
	// 	if !bytes.Equal(block.LastHash(), chain.LastHash()) {
	// 		return false
	// 	}
	// }
	return true
}

func (chain *ChainT) SelectLazy(validators []PubKey) PubKey {
	var (
		finds []PubKey
		diff  = ZeroInt()
	)

	for _, pub := range validators {
		lazyLevel := chain.LazyInterval(pub)

		if lazyLevel.Cmp(diff) > 0 {
			diff = lazyLevel
			finds = []PubKey{pub}
			continue
		}

		if lazyLevel.Cmp(diff) == 0 {
			finds = append(finds, pub)
			continue
		}
	}

	lenpub := uint64(len(finds))

	if lenpub > 1 {
		sort.SliceStable(finds, func(i, j int) bool {
			return strings.Compare(finds[i].Address(), finds[j].Address()) < 0
		})

		rnum := LoadInt(chain.LastHash()).Uint64()
		finds[0] = finds[rnum%lenpub]
	}

	return finds[0]
}

func (chain *ChainT) LazyInterval(pub PubKey) BigInt {
	id := chain.getUserAction(pub)
	if id == nil {
		return ZeroInt()
	}
	return chain.Length().Sub(id)
}

// TODO: LevelDB -> Rollback
func (chain *ChainT) Cut(amount BigInt ) Chain {
	start := chain.Length().Sub(amount)
	end := chain.Length()
	for i:= NewInt(start.String()) ; i.Cmp(end) < 0 ; i.Inc() {
		err := chain.db.Delete([]byte(i.String()) , nil )
		if err!= nil {
			fmt.Println(err)
		}
	}
	chain.setLength(start)
	return chain

	// return &ChainT{
	// 	length: end.Sub(begin),
	// 	blocks: chain.blocks[begin.Uint64():end.Uint64()],
	// }
}
