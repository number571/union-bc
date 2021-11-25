package main

import (
	"fmt"

	"github.com/number571/gopeer/crypto"
	"github.com/number571/laziest/kernel"
)

func main() {
	privs := []crypto.PrivKey{
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
		crypto.NewPrivKey(kernel.KeySize),
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
		kernel.NewTransaction(privs[1], kernel.NewInt("1"), []byte("67890")),
	}

	for _, tx := range txs {
		block.Append(tx)
	}

	block.Accept(privs[0])
	chain.Append(block)

	for _, block := range chain.Blocks() {
		fmt.Println(string(block.Wrap()))
	}

	diff := chain.Interval(privs[2].PubKey())
	fmt.Println(diff)
}
