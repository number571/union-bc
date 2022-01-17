package kernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/number571/go-peer/crypto"
)

var (
	_ Block = &BlockT{}
)

type BlockT struct {
	txs      []Transaction
	prevHash []byte
	currHash []byte
}

type blockJSON struct {
	TXs      [][]byte `json:"txs"`
	PrevHash []byte   `json:"prev_hash"`
	CurrHash []byte   `json:"curr_hash"`
}

func NewBlock(prevHash []byte, txs []Transaction) Block {
	if len(txs) != TXsSize {
		return nil
	}

	for _, tx := range txs {
		if tx == nil {
			fmt.Println("111111111111")
			return nil
		}
		if !tx.IsValid() {
			fmt.Println("222222222222")
			return nil
		}
	}

	sort.SliceStable(txs, func(i, j int) bool {
		return bytes.Compare(txs[i].Hash(), txs[j].Hash()) < 0
	})

	for i := 0; i < len(txs)-1; i++ {
		if bytes.Equal(txs[i].Hash(), txs[i+1].Hash()) {
			fmt.Println("333333333")
			return nil
		}
	}

	block := &BlockT{
		txs:      txs,
		prevHash: prevHash,
	}

	block.currHash = block.newHash()
	return block
}

func LoadBlock(blockBytes []byte) Block {
	blockConv := new(blockJSON)
	err := json.Unmarshal(blockBytes, blockConv)
	if err != nil {
		return nil
	}

	block := &BlockT{
		prevHash: blockConv.PrevHash,
		currHash: blockConv.CurrHash,
	}

	for _, tx := range blockConv.TXs {
		block.txs = append(block.txs, LoadTransaction(tx))
	}

	if !block.IsValid() {
		return nil
	}

	return block
}

func (block *BlockT) Transactions() []Transaction {
	return block.txs
}

func (block *BlockT) PrevHash() Hash {
	return block.prevHash
}

func (block *BlockT) Bytes() []byte {
	blockConv := &blockJSON{
		PrevHash: block.PrevHash(),
		CurrHash: block.Hash(),
	}

	for _, tx := range block.txs {
		blockConv.TXs = append(blockConv.TXs, tx.Bytes())
	}

	blockBytes, err := json.Marshal(blockConv)
	if err != nil {
		return nil
	}

	return blockBytes
}

func (block *BlockT) String() string {
	return fmt.Sprintf("Block{%X}", block.Bytes())
}

func (block *BlockT) Hash() Hash {
	return block.currHash
}

func (block *BlockT) IsValid() bool {
	if len(block.txs) != TXsSize {
		return false
	}

	sort.SliceStable(block.txs, func(i, j int) bool {
		return bytes.Compare(block.txs[i].Hash(), block.txs[j].Hash()) < 0
	})

	return bytes.Equal(block.Hash(), block.newHash())
}

func (block *BlockT) newHash() Hash {
	hash := bytes.Join(
		[][]byte{
			block.PrevHash(),
		},
		[]byte{},
	)

	for _, tx := range block.txs {
		if !tx.IsValid() {
			return nil
		}
		hash = crypto.NewSHA256(bytes.Join(
			[][]byte{
				hash,
				tx.Hash(),
			},
			[]byte{},
		)).Bytes()
	}

	return hash
}
