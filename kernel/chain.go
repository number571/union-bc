package kernel

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/number571/gopeer/crypto"
)

var (
	_ Chain = &ChainT{}
)

type ChainT struct {
	length BigInt
	blocks []Block
}

type chainJSON struct {
	Blocks [][]byte
}

func NewChain(priv crypto.PrivKey, txs []Transaction) Chain {
	chain := &ChainT{
		length: NewInt("0"),
	}

	genesis := NewBlock([]byte(ChainID))
	for _, tx := range txs {
		err := genesis.Append(tx)
		if err != nil {
			return nil
		}
	}

	genesis.Accept(priv)

	err := chain.Append(genesis)
	if err != nil {
		return nil
	}

	return chain
}

func (chain *ChainT) Range(x, y BigInt) Objects {
	return chain.blocks[x.Uint64():y.Uint64()]
}

func (chain *ChainT) Length() BigInt {
	return chain.length
}

func (chain *ChainT) LastHash() Hash {
	last := chain.length.Uint64() - 1
	return chain.blocks[last].Hash()
}

func (chain *ChainT) Append(obj Object) error {
	block := obj.(Block)
	if !block.IsValid() {
		return errors.New("block is invalid")
	}
	chain.blocks = append(chain.blocks, block)
	chain.length = chain.length.Inc()
	return nil
}

func (chain *ChainT) Find(hash Hash) Object {
	for _, block := range chain.blocks {
		if bytes.Equal(hash, block.Hash()) {
			return block
		}
	}
	return nil
}

func (chain *ChainT) IsValid() bool {
	for _, block := range chain.blocks {
		if !block.IsValid() {
			return false
		}
	}
	return true
}

func (chain *ChainT) NonceIsValid(block Block, checkTX Transaction) bool {
	for {
		objects := block.Range(NewInt("0"), chain.Length())
		if objects == nil {
			return false
		}

		for _, tx := range objects.([]Transaction) {
			equalValidator := checkTX.Validator().Address() == tx.Validator().Address()
			if equalValidator && checkTX.Nonce().Cmp(tx.Nonce()) > 0 {
				return true
			}
		}

		blockI := chain.Find(block.PrevHash())
		if blockI == nil {
			return checkTX.Nonce().Cmp(NewInt("0")) == 0
		}

		block = blockI.(Block)
	}
}

func (chain *ChainT) Wrap() []byte {
	chainConv := &chainJSON{}
	for _, block := range chain.blocks {
		chainConv.Blocks = append(chainConv.Blocks, block.Wrap())
	}

	chainBytes, err := json.Marshal(chainConv)
	if err != nil {
		return nil
	}
	return chainBytes
}

func (chain *ChainT) SelectLazy(validators []PubKey) (PubKey, BigInt) {
	var (
		finds []PubKey
		diff  = NewInt("-1")
	)

	for _, pub := range validators {
		lazyLevel := chain.Interval(pub)
		if lazyLevel.Cmp(diff) == 0 {
			finds = append(finds, pub)
			continue
		}

		if lazyLevel.Cmp(diff) > 0 {
			diff = lazyLevel
			finds = []PubKey{pub}
		}
	}

	lenpub := uint64(len(finds))
	if lenpub > 1 {
		hash := chain.LastHash()
		rnum := LoadInt(hash).Uint64()
		return finds[rnum%lenpub], diff
	}

	return finds[0], diff
}

func (chain *ChainT) Interval(pub PubKey) BigInt {
	diff := NewInt("0")
	block := chain.Find(chain.LastHash()).(Block)

	for {
		if pub.Address() == block.Validator().Address() {
			return diff
		}

		objects := block.Range(NewInt("0"), block.Length())
		if objects == nil {
			return NewInt("-1")
		}

		for _, tx := range objects.([]Transaction) {
			if pub.Address() == tx.Validator().Address() {
				return diff
			}
		}

		blockI := chain.Find(block.PrevHash())
		if blockI == nil {
			return NewInt("-1")
		}

		block = blockI.(Block)
		diff = diff.Inc()
	}
}
