package main

import (
	"fmt"

	"github.com/number571/gopeer/crypto"
	"github.com/number571/laziest/kernel"
)

const (
	ChainID = "init-chain"
)

func main() {
	var (
		genegisValidator = crypto.NewPrivKey(kernel.KeySize)

		validators = newPrivKeys()
		valpubs    = newPubKeys(validators)
		txsgen     = newTransactions(validators)
		txs        = newTransactions(newPrivKeys())
	)

	// new genesis block
	genesis := kernel.NewBlock([]byte(ChainID))
	for _, tx := range txsgen {
		genesis.Append(tx)
	}
	genesis.Accept(genegisValidator)

	// new chain
	chain := kernel.NewChain("chain", genesis)

	// append new blocks by PoL
	for i := 0; i < 100; i++ {
		blocks := []kernel.Block{
			newBlock(validators[0], chain, txs),
			newBlock(validators[1], chain, txs),
			newBlock(validators[2], chain, txs),
		}

		// change validator
		validator := chain.SelectLazy(valpubs)
		for _, block := range blocks {
			if validator.Equal(block.Validator()) {
				chain.Append(block)
				// for _, pub := range valpubs {
				// 	fmt.Println(pub.Address(), chain.LazyInterval(pub))
				// }
				// fmt.Println()
				break
			}
		}
	}

	fmt.Println(chain.IsValid())
	fmt.Println(chain.Length())
	pub := genegisValidator.PubKey()
	fmt.Println(pub.Address(), chain.LazyInterval(pub))
	fmt.Println()

	chain.RollBack(kernel.NewInt("90"))

	fmt.Println(chain.IsValid())
	fmt.Println(chain.Length())
	pub = genegisValidator.PubKey()
	fmt.Println(pub.Address(), chain.LazyInterval(pub))
	fmt.Println()

	// append new blocks by PoL
	for i := 0; i < 10; i++ {
		blocks := []kernel.Block{
			newBlock(validators[0], chain, txs),
			newBlock(validators[1], chain, txs),
			newBlock(validators[2], chain, txs),
		}

		// change validator
		validator := chain.SelectLazy(valpubs)
		for _, block := range blocks {
			if validator.Equal(block.Validator()) {
				chain.Append(block)
				break
			}
		}
	}

	fmt.Println(chain.IsValid())
	fmt.Println(chain.Length())
	pub = genegisValidator.PubKey()
	fmt.Println(pub.Address(), chain.LazyInterval(pub))
	fmt.Println()

	// newChain := chain.Snapshot("backup_chain")
	// fmt.Println(newChain.Length())
}

func newPrivKeys() []kernel.PrivKey {
	return []kernel.PrivKey{
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
	}
}

func newPubKeys(privs []kernel.PrivKey) []kernel.PubKey {
	return []kernel.PubKey{
		privs[0].PubKey(),
		privs[1].PubKey(),
		privs[2].PubKey(),
	}
}

func newTransactions(privs []kernel.PrivKey) []kernel.Transaction {
	return []kernel.Transaction{
		kernel.NewTransaction(privs[0], []byte("hello, world!")),
		kernel.NewTransaction(privs[1], []byte("aaabbbccc")),
		kernel.NewTransaction(privs[2], []byte("qwerty")),
	}
}

func newBlock(priv kernel.PrivKey, chain kernel.Chain, txs []kernel.Transaction) kernel.Block {
	block := kernel.NewBlock(chain.LastHash())
	for _, tx := range txs {
		block.Append(tx)
	}
	block.Accept(priv)
	return block
}
