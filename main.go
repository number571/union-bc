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

	txsGenesis := []kernel.Transaction{
		kernel.NewTransaction(privs[0], kernel.NewInt("0"), []byte("hello, world!")),
		kernel.NewTransaction(privs[1], kernel.NewInt("0"), []byte("aaabbbccc")),
		kernel.NewTransaction(privs[1], kernel.NewInt("0"), []byte("qwerty")),
	}

	chain := kernel.NewChain(privs[0], txsGenesis)
	block := kernel.NewBlock(chain.LastHash())

	txs := []kernel.Transaction{
		kernel.NewTransaction(privs[0], kernel.NewInt("1"), []byte("12345")),
		kernel.NewTransaction(privs[2], kernel.NewInt("1"), []byte("67890")),
	}

	for _, tx := range txs {
		block.Append(tx)
	}

	block.Accept(privs[2])
	chain.Append(block)

	list := chain.Range(kernel.NewInt("0"), chain.Length()).([]kernel.Block)
	for _, block := range list {
		fmt.Println(string(block.Wrap()))
	}

	fmt.Println(chain.SelectLazy(pubs))
}
