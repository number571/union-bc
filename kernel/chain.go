package kernel

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	_ Chain = &ChainT{}
)

type ChainT struct {
	state    *leveldb.DB
	journal  *leveldb.DB
	accounts *leveldb.DB
}

func NewChain(path string, genesis Block) Chain {
	var (
		failNotExist = true
		chain        *ChainT
	)

	var (
		statePath    = filepath.Join(path, "state.db")
		journalPath  = filepath.Join(path, "journal.db")
		accountsPath = filepath.Join(path, "accounts.db")
	)

	defer func(chain Chain) {
		if failNotExist {
			return
		}
		chain.Close()
		os.RemoveAll(statePath)
		os.RemoveAll(journalPath)
		os.RemoveAll(accountsPath)
	}(chain)

	if !genesis.IsValid() {
		return nil
	}

	state, err := leveldb.OpenFile(statePath, nil)
	if err != nil {
		failNotExist = false
		return nil
	}

	journal, err := leveldb.OpenFile(journalPath, nil)
	if err != nil {
		failNotExist = false
		return nil
	}

	accounts, err := leveldb.OpenFile(accountsPath, nil)
	if err != nil {
		failNotExist = false
		return nil
	}

	chain = &ChainT{
		state:    state,
		journal:  journal,
		accounts: accounts,
	}

	err = chain.setStateLength(NewInt("0"))
	if err != nil {
		failNotExist = false
		return nil
	}

	err = chain.pushBlock(genesis)
	if err != nil {
		failNotExist = false
		return nil
	}

	return chain
}

func (chain *ChainT) Close() {
	if chain.state != nil {
		chain.state.Close()
	}
	if chain.journal != nil {
		chain.journal.Close()
	}
	if chain.accounts != nil {
		chain.accounts.Close()
	}
}

// range is [x;y]
func (chain *ChainT) Range(x, y BigInt) Object {
	blocks := []Block{}

	if x.Cmp(chain.Length()) > 0 {
		return []Block{}
	}

	for x.Cmp(y) <= 0 {
		block := chain.getStateBlockByID(x)
		if block == nil {
			return blocks
		}
		blocks = append(blocks, block)
		x = x.Inc()
	}

	return blocks
}

func (chain *ChainT) Length() BigInt {
	return chain.getStateLength()
}

func (chain *ChainT) LastHash() Hash {
	return chain.getStateBlockByID(chain.Length()).Hash()
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

	return chain.pushBlock(block)
}

func (chain *ChainT) Find(hash Hash) Object {
	id := chain.getStateBlockIdByHash(hash)
	if id != nil {
		return chain.getStateBlockByID(id)
	}

	id = chain.getJournalBlockIdByTxHash(hash)
	if id != nil {
		return chain.getStateBlockByID(id)
	}

	return nil
}

func (chain *ChainT) IsValid() bool {
	for i := NewInt("1"); i.Cmp(chain.Length()) < 0; i = i.Inc() {
		blocks := chain.Range(i, i.Inc()).([]Block)
		if !blocks[0].IsValid() {
			return false
		}
		if !bytes.Equal(blocks[0].Hash(), blocks[1].LastHash()) {
			return false
		}
	}
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

	if lenpub == 0 {
		panic("length of validators = nil")
	}

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
	lazyHistory := chain.getAccountsLazyByAddress(pub)
	if lazyHistory == nil {
		return ZeroInt()
	}
	return chain.Length().Sub(lazyHistory.last())
}

func (chain *ChainT) RollBack(id BigInt) {
	var (
		mappingPubs = make(map[string]bool)
		startBlock  = chain.Length().Sub(id)
	)

	for i := startBlock.Inc(); i.Cmp(chain.Length()) <= 0; i = i.Inc() {
		block := chain.getStateBlockByID(i)
		if block == nil {
			break
		}

		txs := block.Range(NewInt("1"), block.Length()).([]Transaction)
		for _, tx := range txs {
			pub, addr := tx.Validator(), tx.Validator().Address()
			lazyHistory := chain.getAccountsLazyByAddress(pub)

			if _, ok := mappingPubs[addr]; !ok {
				mappingPubs[addr] = true
				chain.splitAccountsLazyByAddress(pub, startBlock, lazyHistory)
			}

			chain.journal.Delete([]byte(fmt.Sprintf(JournalTxByTxHash, tx.Hash())), nil)
		}

		pub, addr := block.Validator(), block.Validator().Address()
		lazyHistory := chain.getAccountsLazyByAddress(pub)

		if _, ok := mappingPubs[addr]; !ok {
			mappingPubs[addr] = true
			chain.splitAccountsLazyByAddress(pub, startBlock, lazyHistory)
		}

		chain.state.Delete([]byte(fmt.Sprintf(StateBlockByBlockID, i)), nil)
		chain.state.Delete([]byte(fmt.Sprintf(StateBlockIdByBlockHash, block.Hash())), nil)
	}

	chain.setStateLength(chain.Length().Sub(id))
}
