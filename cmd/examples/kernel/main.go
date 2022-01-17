package main

import (
	"fmt"
	"os"

	"github.com/number571/gopeer/crypto"
	"github.com/number571/unionbc/kernel"
)

func main() {
	var chain kernel.Chain
	chainPath := "chain"

	priv := crypto.NewPrivKey(kernel.KeySize)

	if _, err := os.Stat(chainPath); os.IsNotExist(err) {
		genesis := kernel.NewBlock(
			[]byte("genesis.block"),
			[]kernel.Transaction{
				kernel.NewTransaction(priv, []byte("info-G")),
				kernel.NewTransaction(priv, []byte("info-G")),
			},
		)
		chain = kernel.NewChain(chainPath, genesis)
	} else {
		chain = kernel.LoadChain(chainPath)
	}

	for i := 0; i < 5; i++ {
		block := kernel.NewBlock(
			chain.Block(chain.Height()).Hash(),
			[]kernel.Transaction{
				kernel.NewTransaction(priv, []byte(fmt.Sprintf("info-%d", i))),
				kernel.NewTransaction(priv, []byte(fmt.Sprintf("info-%d", i))),
			},
		)
		chain.Accept(block)
	}

	for h := kernel.Height(0); h <= chain.Height(); h++ {
		block := chain.Block(h)
		for _, tx := range block.Transactions() {
			fmt.Printf("TX: {validator: %s, pay_load: %s}\n",
				tx.Validator().Address(),
				string(tx.PayLoad()))
		}
	}
}
