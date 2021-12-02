package kernel

import (
	"errors"
	"fmt"
)

const (
	StateLength      = "state.length"
	StateBlockByID   = "state.block[block_id=%s]"
	StateBlockByHash = "state.block[block_hash=%s]"

	JournalTxByHash        = "journal.tx[tx_hash=%s]"
	JournalBlockIdByTxHash = "journal.block_id[tx_hash=%s]"

	// TODO: Int[Lazy] -> []{Int[Block], Int[Lazy]}
	AccountsLazyByAddress = "accounts.lazy[address=%s]"
)

// Length

func (chain *ChainT) setStateLength(length BigInt) error {
	return chain.state.Put([]byte(StateLength), length.Bytes(), nil)
}

func (chain *ChainT) getStateLength() BigInt {
	data, err := chain.state.Get([]byte(StateLength), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}

// Blocks

func (chain *ChainT) pushBlock(block Block) error {
	type backLazyByAddressT struct {
		pub      PubKey
		interval BigInt
	}

	var (
		failNotExist = true
	)

	var (
		backTxByHash      []Hash
		backLazyByAddress []backLazyByAddressT
	)

	length := chain.getStateLength()
	if length == nil {
		return errors.New("length is nil")
	}

	defer func() {
		if failNotExist {
			return
		}

		chain.setStateLength(length)
		newLength := length.Inc()

		chain.state.Delete([]byte(fmt.Sprintf(StateBlockByID, newLength)), nil)
		chain.state.Delete([]byte(fmt.Sprintf(StateBlockByHash, block.Hash())), nil)

		for _, hash := range backTxByHash {
			chain.journal.Delete([]byte(fmt.Sprintf(JournalTxByHash, hash)), nil)
		}

		for _, lazy := range backLazyByAddress {
			chain.setLazyByAddress(lazy.pub, lazy.interval)
		}
	}()

	newLength := length.Inc()
	err := chain.setStateLength(newLength)
	if err != nil {
		failNotExist = false
		return err
	}

	err = chain.setStateBlockByID(newLength, block)
	if err != nil {
		failNotExist = false
		return err
	}

	err = chain.setStateBlockByHash(block.Hash(), block)
	if err != nil {
		failNotExist = false
		return err
	}

	err = chain.setLazyByAddress(block.Validator(), newLength)
	if err != nil {
		failNotExist = false
		return err
	}

	txs := block.Range(ZeroInt(), block.Length()).([]Transaction)
	if txs == nil {
		failNotExist = false
		return errors.New("txs is nil")
	}

	for _, tx := range txs {
		backTxByHash = append(backTxByHash, tx.Hash())
		err = chain.setTxByHash(tx.Hash(), tx)
		if err != nil {
			failNotExist = false
			return err
		}

		err = chain.setBlockIdByTxHash(tx.Hash(), newLength)
		if err != nil {
			failNotExist = false
			return err
		}

		pub := tx.Validator()
		backLazyByAddress = append(backLazyByAddress, backLazyByAddressT{
			pub, chain.getLazyByAddress(pub)})

		err = chain.setLazyByAddress(pub, newLength)
		if err != nil {
			failNotExist = false
			return err
		}
	}

	return nil
}

func (chain *ChainT) setStateBlockByID(id BigInt, block Block) error {
	return chain.state.Put(
		[]byte(fmt.Sprintf(StateBlockByID, id)),
		block.Wrap(), nil)
}

func (chain *ChainT) setStateBlockByHash(hash Hash, block Block) error {
	return chain.state.Put(
		[]byte(fmt.Sprintf(StateBlockByHash, hash)),
		block.Wrap(), nil)
}

func (chain *ChainT) getBlockByID(id BigInt) Block {
	data, err := chain.state.Get([]byte(fmt.Sprintf(StateBlockByID, id)), nil)
	if err != nil {
		return nil
	}
	return LoadBlock(data)
}

func (chain *ChainT) getBlockByHash(hash Hash) Block {
	data, err := chain.state.Get([]byte(fmt.Sprintf(StateBlockByHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadBlock(data)
}

// Transactions

func (chain *ChainT) setTxByHash(hash Hash, tx Transaction) error {
	return chain.journal.Put(
		[]byte(fmt.Sprintf(JournalTxByHash, hash)),
		tx.Wrap(), nil)
}

func (chain *ChainT) getTxByHash(hash Hash) Transaction {
	data, err := chain.journal.Get([]byte(fmt.Sprintf(JournalTxByHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadTransaction(data)
}

func (chain *ChainT) setBlockIdByTxHash(hash Hash, id BigInt) error {
	return chain.journal.Put(
		[]byte(fmt.Sprintf(JournalBlockIdByTxHash, hash)),
		id.Bytes(), nil)
}

func (chain *ChainT) getBlockIdByTxHash(hash Hash) BigInt {
	data, err := chain.journal.Get([]byte(fmt.Sprintf(JournalBlockIdByTxHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}

// Users

func (chain *ChainT) setLazyByAddress(pub PubKey, id BigInt) error {
	return chain.accounts.Put(
		[]byte(fmt.Sprintf(AccountsLazyByAddress, pub.Address())),
		id.Bytes(), nil)
}

func (chain *ChainT) getLazyByAddress(pub PubKey) BigInt {
	data, err := chain.accounts.Get([]byte(fmt.Sprintf(AccountsLazyByAddress, pub.Address())), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}
