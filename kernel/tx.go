package kernel

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/number571/gopeer/crypto"
)

var (
	_ Transaction = &TransactionT{}
)

type TransactionT struct {
	payLoad   []byte
	hash      []byte
	sign      []byte
	validator crypto.PubKey
}

type txJSON struct {
	PayLoad   []byte `json:"pay_load"`
	Hash      []byte `json:"hash"`
	Sign      []byte `json:"sign"`
	Validator []byte `json:"validator"`
}

func NewTransaction(priv PrivKey, payLoad []byte) Transaction {
	if priv == nil {
		return nil
	}

	if priv.Size() != KeySize {
		return nil
	}

	tx := &TransactionT{
		payLoad:   payLoad,
		validator: priv.PubKey(),
	}

	tx.hash = tx.newHash()
	tx.sign = priv.Sign(tx.hash)

	return tx
}

func LoadTransaction(txbytes []byte) Transaction {
	txConv := new(txJSON)
	err := json.Unmarshal(txbytes, txConv)
	if err != nil {
		return nil
	}

	tx := &TransactionT{
		payLoad:   txConv.PayLoad,
		hash:      txConv.Hash,
		sign:      txConv.Sign,
		validator: crypto.LoadPubKey(txConv.Validator),
	}

	if !tx.IsValid() {
		return nil
	}

	return tx
}

func (tx *TransactionT) PayLoad() []byte {
	return tx.payLoad
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

func (tx *TransactionT) Bytes() []byte {
	txConv := &txJSON{
		PayLoad:   tx.PayLoad(),
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

func (tx *TransactionT) String() string {
	return fmt.Sprintf("TX{%X}", tx.Bytes())
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
			tx.PayLoad(),
		},
		[]byte{},
	)).Bytes()
}
