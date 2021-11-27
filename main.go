package main

import (
	"fmt"

	"github.com/number571/gopeer/crypto"
	"github.com/number571/laziest/kernel"
)

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

func main() {
	var (
		validators = newPrivKeys()
		valpubs    = newPubKeys(validators)
		txsgen     = newTransactions(validators)
		txs        = newTransactions(newPrivKeys())
	)

	chain := kernel.NewChain(validators[0], txsgen)

	for i := 0; i < 100; i++ {
		blocks := []kernel.Block{
			newBlock(validators[0], chain, txs),
			newBlock(validators[1], chain, txs),
			newBlock(validators[2], chain, txs),
		}

		validator := chain.SelectLazy(valpubs)
		for _, block := range blocks {
			if validator.Equal(block.Validator()) {
				chain.Append(block)
				break
			}
		}
	}

	begin := kernel.NewInt("0")
	end := chain.Length()

	list := chain.Range(begin, end).([]kernel.Block)
	for _, block := range list {
		fmt.Println(block.Validator().Address())
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
