package kernel

import (
	"errors"
	"fmt"
)

const (
	KeyLength          = "chain.length"
	KeyBlockID         = "chain.block.id.%s"
	KeyBlockHash       = "chain.block.hash.%s"
	KeyTransactionHash = "chain.tx.hash.%s"
	KeyUserAction      = "chain.user.action.%s"
)

// Length

func (chain *ChainT) setLength(length BigInt) error {
	return chain.setLevelDB(KeyLength, length.Bytes())
}

func (chain *ChainT) getLength() BigInt {
	data := chain.getLevelDB(KeyLength)
	if data == nil {
		return nil
	}
	return LoadInt(data)
}

// Blocks

func (chain *ChainT) setBlock(block Block) error {
	length := chain.getLength()
	if length == nil {
		return errors.New("length is nil")
	}

	newLength := length.Inc()

	err := chain.setLength(newLength)
	if err != nil {
		return err
	}

	// NEED ATOMIC!!!
	err = chain.setLevelDB(fmt.Sprintf(KeyBlockID, newLength), block.Wrap())
	if err != nil {
		return err
	}

	// NEED ATOMIC!!!
	err = chain.setLevelDB(fmt.Sprintf(KeyUserAction, block.Validator().Bytes()), newLength.Bytes())
	if err != nil {
		return err
	}

	// NEED ATOMIC!!!
	err = chain.setLevelDB(fmt.Sprintf(KeyBlockHash, block.Hash()), block.Wrap())
	if err != nil {
		return err
	}

	begin := ZeroInt()
	end := block.Length()
	txs := block.Range(begin, end).([]Transaction)
	if txs == nil {
		return errors.New("txs is nil")
	}

	// NEED ATOMIC!!!
	for _, tx := range txs {
		err = chain.setTransaction(tx)
		if err != nil {
			return err
		}
		err = chain.setLevelDB(fmt.Sprintf(KeyUserAction, tx.Validator().Bytes()), newLength.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (chain *ChainT) getBlockByID(id BigInt) Block {
	data := chain.getLevelDB(fmt.Sprintf(KeyBlockID, id))
	if data == nil {
		return nil
	}
	return LoadBlock(data)
}

func (chain *ChainT) getBlockByHash(hash Hash) Block {
	data := chain.getLevelDB(fmt.Sprintf(KeyBlockHash, hash))
	if data == nil {
		return nil
	}
	return LoadBlock(data)
}

// Transactions

func (chain *ChainT) setTransaction(tx Transaction) error {
	return chain.setLevelDB(fmt.Sprintf(KeyTransactionHash, tx.Hash()), tx.Wrap())
}

func (chain *ChainT) getTransaction(hash Hash) Transaction {
	data := chain.getLevelDB(fmt.Sprintf(KeyTransactionHash, hash))
	if data == nil {
		return nil
	}
	return LoadTransaction(data)
}

// Users

func (chain *ChainT) getUserAction(pub PubKey) BigInt {
	data := chain.getLevelDB(fmt.Sprintf(KeyUserAction, pub.Bytes()))
	if data == nil {
		return nil
	}
	return LoadInt(data)
}

// LevelDB

func (chain *ChainT) getLevelDB(key string) []byte {
	data, err := chain.db.Get([]byte(key), nil)
	if err != nil {
		return nil
	}
	return data
}

func (chain *ChainT) setLevelDB(key string, value []byte) error {
	return chain.db.Put([]byte(key), value, nil)
}
