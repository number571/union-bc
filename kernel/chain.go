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
	blocks []Block
}

type chainJSON struct {
	Blocks [][]byte
}

func NewChain(priv crypto.PrivKey, txs []Transaction) Chain {
	chain := &ChainT{}

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

func (chain *ChainT) Blocks() []Block {
	return chain.blocks
}

func (chain *ChainT) LastHash() Hash {
	last := len(chain.blocks) - 1
	return chain.blocks[last].Hash()
}

func (chain *ChainT) Append(obj Object) error {
	block := obj.(Block)
	if !block.IsValid() {
		return errors.New("block is invalid")
	}
	chain.blocks = append(chain.blocks, block)
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
		for _, tx := range block.TXs() {
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

func (chain *ChainT) Interval(pub PubKey) BigInt {
	diff := NewInt("0")
	block := chain.Find(chain.LastHash()).(Block)

	for {
		if pub.Address() == block.Validator().Address() {
			return diff
		}

		for _, tx := range block.TXs() {
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
