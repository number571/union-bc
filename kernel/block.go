package kernel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/number571/gopeer/crypto"
)

var (
	_ Block = &BlockT{}
)

type BlockT struct {
	accepted  bool
	txs       []Transaction
	prevHash  []byte
	currHash  []byte
	sign      []byte
	validator crypto.PubKey
}

type blockJSON struct {
	TXs       [][]byte `json:"txs"`
	PrevHash  []byte   `json:"prev_hash"`
	CurrHash  []byte   `json:"curr_hash"`
	Sign      []byte   `json:"sign"`
	Validator []byte   `json:"validator"`
}

func NewBlock(prevHash []byte) Block {
	return &BlockT{
		prevHash: prevHash,
	}
}

func LoadBlock(blockBytes []byte) Block {
	blockConv := new(blockJSON)
	json.Unmarshal(blockBytes, blockConv)

	block := &BlockT{
		accepted:  true,
		prevHash:  blockConv.PrevHash,
		currHash:  blockConv.CurrHash,
		sign:      blockConv.Sign,
		validator: crypto.LoadPubKey(blockConv.Validator),
	}

	for _, tx := range blockConv.TXs {
		block.txs = append(block.txs, LoadTransaction(tx))
	}

	if !block.IsValid() {
		return nil
	}

	return block
}

// range is [x;y]
func (block *BlockT) Range(x, y BigInt) Object {
	return block.txs[x.Dec().Uint64():y.Uint64()]
}

func (block *BlockT) Length() BigInt {
	return NewInt(fmt.Sprintf("%d", len(block.txs)))
}

func (block *BlockT) LastHash() Hash {
	return block.prevHash
}

func (block *BlockT) Append(obj Object) error {
	tx := obj.(Transaction)
	if tx == nil {
		return errors.New("tx is nil")
	}

	if !tx.IsValid() {
		return errors.New("tx is invalid")
	}

	block.txs = append(block.txs, tx)
	return nil
}

func (block *BlockT) Find(hash Hash) Object {
	for _, tx := range block.txs {
		if bytes.Equal(hash, tx.Hash()) {
			return tx
		}
	}
	return nil
}

func (block *BlockT) Accept(priv PrivKey) error {
	if priv == nil {
		return errors.New("priv is nil")
	}

	if priv.Size() != KeySize {
		return errors.New("key size not allowed")
	}

	sort.SliceStable(block.txs, func(i, j int) bool {
		return bytes.Compare(block.txs[i].Hash(), block.txs[j].Hash()) < 0
	})

	block.accepted = true
	block.validator = priv.PubKey()
	block.currHash = block.newHash()
	block.sign = priv.Sign(block.currHash)

	return nil
}

func (block *BlockT) Wrap() []byte {
	if !block.accepted {
		return nil
	}

	blockConv := &blockJSON{
		PrevHash:  block.LastHash(),
		CurrHash:  block.Hash(),
		Sign:      block.Sign(),
		Validator: block.validator.Bytes(),
	}

	for _, tx := range block.txs {
		blockConv.TXs = append(blockConv.TXs, tx.Wrap())
	}

	blockBytes, err := json.Marshal(blockConv)
	if err != nil {
		return nil
	}

	return blockBytes
}

func (block *BlockT) Hash() Hash {
	return block.currHash
}

func (block *BlockT) Sign() Sign {
	return block.sign
}

func (block *BlockT) Validator() PubKey {
	return block.validator
}

func (block *BlockT) IsValid() bool {
	if block.Validator() == nil {
		return false
	}

	sort.SliceStable(block.txs, func(i, j int) bool {
		return bytes.Compare(block.txs[i].Hash(), block.txs[j].Hash()) < 0
	})

	if !bytes.Equal(block.Hash(), block.newHash()) {
		return false
	}

	return block.Validator().Verify(block.Hash(), block.Sign())
}

func (block *BlockT) newHash() Hash {
	hash := bytes.Join(
		[][]byte{
			block.Validator().Bytes(),
			block.LastHash(),
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
