package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/number571/go-peer/crypto"
	"github.com/number571/go-peer/encoding"
	"github.com/number571/union-bc/kernel"
	"github.com/number571/union-bc/network"
)

var (
	Mutex sync.Mutex
	Chain kernel.Chain
)

type updateBlock struct {
	Height kernel.Height `json:"height"`
	Block  []byte        `json:"block"`
}

func init() {
	var (
		chainPath = "chain" + os.Args[1]
	)

	if pathIsExist(chainPath) {
		Chain = kernel.LoadChain(chainPath)
	} else {
		Chain = kernel.NewChain(chainPath, newGenesis())
	}
}

func main() {
	var (
		moniker  = "init-moniker"
		address  = os.Args[1]
		listAddr = []string{
			":7070",
			":8080",
			":9090",
		}
	)

	node := network.NewNode(moniker).
		Handle(MsgGetHeight, handleGetHeight).
		Handle(MsgGetBlock, handleGetBlock).
		Handle(MsgSetBlock, handleSetBlock).
		Handle(MsgGetTX, handleGetTX).
		Handle(MsgSetTX, handleSetTX)

	for _, addr := range listAddr {
		if addr == address {
			continue
		}
		node.Connect(addr)
	}

	fmt.Println("Node is listening...")
	if address != "" {
		go node.Listen(address)
	}

	if Chain.Height() == 0 {
		upBlock := updateBlock{
			Height: 0,
			Block:  Chain.Block(0).Bytes(),
		}
		upBlockBytes, err := json.Marshal(upBlock)
		if err != nil {
			return
		}
		node.Broadcast(network.NewMessage(MsgSetBlock, upBlockBytes))
	}

	for {
		command := inputString("> ")
		splited := strings.Split(command, " ")

		switch splited[0] {
		case "bc":
			for i := kernel.Height(0); i <= Chain.Height(); i++ {
				block := Chain.Block(i)

				fmt.Println("Block:", i)
				fmt.Printf("Previous Hash: %X\n", block.PrevHash())
				fmt.Printf("Current Hash: %X\n", block.Hash())
				for _, tx := range block.Transactions() {
					fmt.Printf("| TX: {pay_load: %s}\n", string(tx.PayLoad()))
				}

				fmt.Println()
			}

		case "tx":
			msg := []byte(strings.Join(splited[1:], " "))
			tx := kernel.NewTransaction(crypto.NewPrivKey(kernel.KeySize), msg)
			node.Broadcast(network.NewMessage(MsgSetTX, tx.Bytes()))

		case "net":
			fmt.Println(len(node.Connections()))

		case "mp":
			fmt.Println(Chain.Mempool().Height())
		}
	}
}

func handleGetHeight(node network.Node, conn network.Conn, data []byte) {
	Mutex.Lock()
	defer Mutex.Unlock()

	var (
		height = uint64(Chain.Height())
	)

	msg := network.NewMessage(
		MsgGetHeight|MaskBit,
		encoding.Uint64ToBytes(height),
	)

	conn.Write(msg.Bytes())
}

func handleGetBlock(node network.Node, conn network.Conn, data []byte) {
	Mutex.Lock()
	defer Mutex.Unlock()

	var (
		height     = kernel.Height(encoding.BytesToUint64(data))
		block      = Chain.Block(height)
		blockBytes = []byte{}
	)

	if block != nil {
		blockBytes = block.Bytes()
	}

	msg := network.NewMessage(
		MsgGetBlock|MaskBit,
		blockBytes,
	)

	conn.Write(msg.Bytes())
}

func handleSetBlock(node network.Node, conn network.Conn, data []byte) {
	Mutex.Lock()
	defer Mutex.Unlock()

	var (
		upBlock   updateBlock
		mempool   = Chain.Mempool()
		height    = Chain.Height()
		lastBlock = Chain.Block(height)
	)

	err := json.Unmarshal(data, &upBlock)
	if err != nil {
		return
	}

	block := kernel.LoadBlock(upBlock.Block)
	if block == nil {
		return
	}

	if upBlock.Height < height {
		return
	}

	if upBlock.Height > height+1 {
		return
	}

	if upBlock.Height == height+1 {
		for _, tx := range block.Transactions() {
			mempool.Push(tx)
		}

		tryUpdateBlock(node, mempool)
		return
	}

	// if upBlock.Height == height
	// then try merge block

	if bytes.Equal(block.Hash(), lastBlock.Hash()) {
		return
	}

	ok := Chain.Merge(block.Transactions())
	if !ok {
		return
	}

	upBlock = updateBlock{
		Height: height,
		Block:  Chain.Block(height).Bytes(),
	}
	upBlockBytes, err := json.Marshal(upBlock)
	if err != nil {
		return
	}

	node.Broadcast(network.NewMessage(MsgSetBlock, upBlockBytes))
}

func handleGetTX(node network.Node, conn network.Conn, data []byte) {
	Mutex.Lock()
	defer Mutex.Unlock()

	var (
		hash    = kernel.Hash(data)
		tx      = Chain.TX(hash)
		txBytes = []byte{}
	)

	if tx != nil {
		txBytes = tx.Bytes()
	}

	msg := network.NewMessage(
		MsgGetTX|MaskBit,
		txBytes,
	)

	conn.Write(msg.Bytes())
}

func handleSetTX(node network.Node, conn network.Conn, data []byte) {
	Mutex.Lock()
	defer Mutex.Unlock()

	var (
		mempool = Chain.Mempool()
		tx      = kernel.LoadTransaction(data)
		hash    = tx.Hash()
	)

	if tx == nil {
		return
	}

	txInChain := Chain.TX(hash)
	if txInChain != nil {
		return
	}

	txInMempool := mempool.TX(hash)
	if txInMempool != nil {
		return
	}

	mempool.Push(tx)

	// node.Broadcast(network.NewMessage(MsgSetTX, data))

	if mempool.Height() < kernel.TXsSize {
		return
	}

	tryUpdateBlock(node, mempool)
}

func tryUpdateBlock(node network.Node, mempool kernel.Mempool) {
	txs := []kernel.Transaction{}

	for i := 0; i < kernel.TXsSize; i++ {
		tx := mempool.Pop()
		if tx == nil {
			return
		}
		txs = append(txs, tx)
	}

	height := Chain.Height()
	lastBlock := Chain.Block(height)
	newBlock := kernel.NewBlock(lastBlock.Hash(), txs)

	ok := Chain.Accept(newBlock)
	if !ok {
		return
	}

	upBlock := updateBlock{
		Height: height + 1,
		Block:  newBlock.Bytes(),
	}

	upBlockBytes, err := json.Marshal(upBlock)
	if err != nil {
		return
	}

	node.Broadcast(network.NewMessage(MsgSetBlock, upBlockBytes))
}

func inputString(begin string) string {
	fmt.Print(begin)
	s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(s, "\n", "", -1)
}

func pathIsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func newGenesis() kernel.Block {
	var (
		priv = crypto.NewPrivKey(kernel.KeySize)
		txs  = []kernel.Transaction{}
	)

	for i := 0; i < kernel.TXsSize; i++ {
		data := []byte(fmt.Sprintf("info-G-%d", i))
		txs = append(txs, kernel.NewTransaction(priv, data))
	}
	return kernel.NewBlock(
		[]byte("genesis.block"),
		txs,
	)
}
