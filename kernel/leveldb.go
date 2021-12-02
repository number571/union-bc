package kernel

import (
	"errors"
	"fmt"
)

const (
	StateLength             = "state.length"
	StateBlockByBlockID     = "state.block[block_id=%s]"
	StateBlockIdByBlockHash = "state.block_id[block_hash=%s]"

	JournalTxByTxHash      = "journal.tx[tx_hash=%s]"
	JournalBlockIdByTxHash = "journal.block_id[tx_hash=%s]"

	// TODO: Int[Lazy] -> []{Int[Block], Int[Lazy]}
	AccountsLazyByAddress = "accounts.lazy[address=%s]"
)

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

		chain.state.Delete([]byte(fmt.Sprintf(StateBlockByBlockID, newLength)), nil)
		chain.state.Delete([]byte(fmt.Sprintf(StateBlockIdByBlockHash, block.Hash())), nil)

		for _, hash := range backTxByHash {
			chain.journal.Delete([]byte(fmt.Sprintf(JournalTxByTxHash, hash)), nil)
		}

		for _, lazy := range backLazyByAddress {
			chain.setAccountsLazyByAddress(lazy.pub, lazy.interval)
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

	err = chain.setStateBlockIdByHash(block.Hash(), newLength)
	if err != nil {
		failNotExist = false
		return err
	}

	err = chain.setAccountsLazyByAddress(block.Validator(), newLength)
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
		err = chain.setJournalTxByTxHash(tx.Hash(), tx)
		if err != nil {
			failNotExist = false
			return err
		}

		err = chain.setJournalBlockIdByTxHash(tx.Hash(), newLength)
		if err != nil {
			failNotExist = false
			return err
		}

		pub := tx.Validator()
		backLazyByAddress = append(backLazyByAddress, backLazyByAddressT{
			pub, chain.getAccountsLazyByAddress(pub)})

		err = chain.setAccountsLazyByAddress(pub, newLength)
		if err != nil {
			failNotExist = false
			return err
		}
	}

	return nil
}

func (chain *ChainT) setStateBlockByID(id BigInt, block Block) error {
	return chain.state.Put(
		[]byte(fmt.Sprintf(StateBlockByBlockID, id)),
		block.Wrap(), nil)
}

func (chain *ChainT) getStateBlockByID(id BigInt) Block {
	data, err := chain.state.Get([]byte(fmt.Sprintf(StateBlockByBlockID, id)), nil)
	if err != nil {
		return nil
	}
	return LoadBlock(data)
}

func (chain *ChainT) setStateBlockIdByHash(hash Hash, id BigInt) error {
	return chain.state.Put(
		[]byte(fmt.Sprintf(StateBlockIdByBlockHash, hash)),
		id.Bytes(), nil)
}

func (chain *ChainT) getStateBlockIdByHash(hash Hash) BigInt {
	data, err := chain.state.Get([]byte(fmt.Sprintf(StateBlockIdByBlockHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}

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

// Transactions

func (chain *ChainT) setJournalTxByTxHash(hash Hash, tx Transaction) error {
	return chain.journal.Put(
		[]byte(fmt.Sprintf(JournalTxByTxHash, hash)),
		tx.Wrap(), nil)
}

func (chain *ChainT) getJournalTxByTxHash(hash Hash) Transaction {
	data, err := chain.journal.Get([]byte(fmt.Sprintf(JournalTxByTxHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadTransaction(data)
}

func (chain *ChainT) setJournalBlockIdByTxHash(hash Hash, id BigInt) error {
	return chain.journal.Put(
		[]byte(fmt.Sprintf(JournalBlockIdByTxHash, hash)),
		id.Bytes(), nil)
}

func (chain *ChainT) getJournalBlockIdByTxHash(hash Hash) BigInt {
	data, err := chain.journal.Get([]byte(fmt.Sprintf(JournalBlockIdByTxHash, hash)), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}

// Users

func (chain *ChainT) setAccountsLazyByAddress(pub PubKey, id BigInt) error {
	return chain.accounts.Put(
		[]byte(fmt.Sprintf(AccountsLazyByAddress, pub.Address())),
		id.Bytes(), nil)
}

func (chain *ChainT) getAccountsLazyByAddress(pub PubKey) BigInt {
	data, err := chain.accounts.Get([]byte(fmt.Sprintf(AccountsLazyByAddress, pub.Address())), nil)
	if err != nil {
		return nil
	}
	return LoadInt(data)
}
