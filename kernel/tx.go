package kernel

import (
	"bytes"
	"encoding/json"

	"github.com/number571/gopeer/crypto"
)

var (
	_ Transaction = &TransactionT{}
)

type TransactionT struct {
	nonce     []byte
	data      []byte
	hash      []byte
	sign      []byte
	validator crypto.PubKey
}

type txJSON struct {
	Nonce     []byte `json:"nonce"`
	Data      []byte `json:"data"`
	Hash      []byte `json:"hash"`
	Sign      []byte `json:"sign"`
	Validator []byte `json:"validator"`
}

func NewTransaction(priv PrivKey, nonce BigInt, data []byte) Transaction {
	if priv == nil {
		return nil
	}

	if priv.Size() != KeySize {
		return nil
	}

	tx := &TransactionT{
		nonce:     nonce.Bytes(),
		data:      data,
		validator: priv.PubKey(),
	}

	tx.hash = tx.newHash()
	tx.sign = priv.Sign(tx.hash)

	return tx
}

func LoadTransaction(txbytes []byte) Transaction {
	txConv := new(txJSON)
	json.Unmarshal(txbytes, txConv)

	tx := &TransactionT{
		nonce:     txConv.Nonce,
		data:      txConv.Data,
		hash:      txConv.Hash,
		sign:      txConv.Sign,
		validator: crypto.LoadPubKey(txConv.Validator),
	}

	if !tx.IsValid() {
		return nil
	}

	return tx
}

func (tx *TransactionT) Nonce() BigInt {
	return LoadInt(tx.nonce)
}

func (tx *TransactionT) Data() []byte {
	return tx.data
}

func (tx *TransactionT) Hash() Hash {
	return tx.hash
}

func (tx *TransactionT) Sign() Sign {
	return tx.sign
}

func (tx *TransactionT) Validator() PubKey {
	return tx.validator
}

func (tx *TransactionT) Wrap() []byte {
	txConv := &txJSON{
		Nonce:     tx.Nonce().Bytes(),
		Data:      tx.Data(),
		Hash:      tx.Hash(),
		Sign:      tx.Sign(),
		Validator: tx.Validator().Bytes(),
	}
	txbytes, err := json.Marshal(txConv)
	if err != nil {
		return nil
	}
	return txbytes
}

func (tx *TransactionT) IsValid() bool {
	if tx.Validator() == nil {
		return false
	}

	if !bytes.Equal(tx.Hash(), tx.newHash()) {
		return false
	}

	return tx.Validator().Verify(tx.Hash(), tx.Sign())
}

func (tx *TransactionT) newHash() Hash {
	return crypto.NewSHA256(bytes.Join(
		[][]byte{
			tx.Validator().Bytes(),
			tx.Nonce().Bytes(),
			tx.Data(),
		},
		[]byte{},
	)).Bytes()
}
