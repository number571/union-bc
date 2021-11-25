package main

import (
	"fmt"

	"github.com/number571/gopeer/crypto"
	"github.com/number571/laziest/kernel"
)

func main() {
	privs := []kernel.PrivKey{
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
	}

	pubs := []kernel.PubKey{
		privs[0].PubKey(),
		privs[1].PubKey(),
		privs[2].PubKey(),
	}

	txs := []kernel.Transaction{
		kernel.NewTransaction(privs[0], kernel.NewInt("0"), []byte("hello, world!")),
		kernel.NewTransaction(privs[1], kernel.NewInt("0"), []byte("aaabbbccc")),
		kernel.NewTransaction(privs[2], kernel.NewInt("0"), []byte("qwerty")),
	}

	chain := kernel.NewChain(privs[0], txs)

	for i := 0; i < 100; i++ {
		blocks := []kernel.Block{
			newBlock(privs[0], chain, txs),
			newBlock(privs[1], chain, txs),
			newBlock(privs[2], chain, txs),
		}

		validator := chain.SelectLazy(pubs)
		for _, block := range blocks {
			if validator.Address() == block.Validator().Address() {
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
