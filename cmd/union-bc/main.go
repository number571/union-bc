package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/number571/go-peer/crypto"
	"github.com/number571/go-peer/encoding"
	"github.com/number571/union-bc/kernel"
	"github.com/number571/union-bc/network"
)

var (
	Chain       kernel.Chain
	CurrentTime uint64
	ChainPath   = "chain" + os.Args[1]
)

var (
	Address  = os.Args[1]
	ListAddr = []string{
		":7070",
		":8080",
		":9090",
	}
)

type updateBlock struct {
	Height kernel.Height `json:"height"`
	Block  []byte        `json:"block"`
}

func init() {
	if pathIsExist(ChainPath) {
		Chain = kernel.LoadChain(ChainPath)
	} else {
		Chain = kernel.NewChain(ChainPath, newGenesis())
	}
}

func main() {
	node := network.NewNode("init-moniker").
		Handle(MsgGetTime, handleGetTime).
		Handle(MsgGetHeight, handleGetHeight).
		Handle(MsgGetBlock, handleGetBlock).
		Handle(MsgSetBlock, handleSetBlock).
		Handle(MsgGetTX, handleGetTX).
		Handle(MsgSetTX, handleSetTX)

	initNode(node)
	initClient()
}

func initClient() {
	time.Sleep(1 * time.Second)

	client := network.NewClient(Address)
	if client == nil {
		panic("client is nil")
	}

	for {
		priv := crypto.NewPrivKey(kernel.KeySize)
		for i := 0; i < 20; i++ {
			tx := kernel.NewTransaction(priv, []byte(crypto.RandString(20)))
			_ = client.Request(network.NewMessage(MsgSetTX, tx.Bytes()))
		}
		time.Sleep(3 * time.Second)
	}
}

func initNode(node network.Node) {
	fmt.Println("Node is listening...")
	var client network.Client

	for _, addr := range ListAddr {
		if addr == Address {
			continue
		}
		client = network.NewClient(addr)
		if client == nil {
			continue
		}
		break
	}

	if client != nil {
		syncBlocks(client)
		atomic.StoreUint64(&CurrentTime, getTime(client))
		client.Close()
	}

	// Connects
	for _, addr := range ListAddr {
		if addr == Address {
			continue
		}
		node.Connect(addr)
	}

	// Listen port
	go node.Listen(Address)

	// Run timer
	go func(node network.Node) {
		for {
			time.Sleep(1 * time.Second)
			atomic.AddUint64(&CurrentTime, 1)

			if CurrentTime%IntervalTime == 0 {
				node.Lock()
				tryUpdateBlock(node, Chain.Mempool(), Chain.Height())
				node.Unlock()
			}
		}
	}(node)
}

func getHeight(client network.Client) kernel.Height {
	msg := network.NewMessage(
		MsgGetHeight,
		nil,
	)

	msg = client.Request(msg)
	if msg == nil {
		return 0
	}

	return kernel.Height(encoding.BytesToUint64(msg.Body()))
}

func getBlock(client network.Client, height kernel.Height) kernel.Block {
	msg := network.NewMessage(
		MsgGetBlock,
		encoding.Uint64ToBytes(uint64(height)),
	)
	msg = client.Request(msg)
	if msg == nil {
		return nil
	}
	return kernel.LoadBlock(msg.Body())
}

func getTime(client network.Client) uint64 {
	msg := network.NewMessage(
		MsgGetTime,
		nil,
	)

	msg = client.Request(msg)
	if msg == nil {
		return 0
	}

	return encoding.BytesToUint64(msg.Body())
}

func syncBlocks(client network.Client) {
	mempool := Chain.Mempool()

	initHeight := Chain.Height()
	limitHeight := getHeight(client)

	// syncable blocks
	for i := initHeight; i <= limitHeight; i++ {
		block := getBlock(client, i)
		if block == nil {
			break
		}

		if i == 0 {
			Chain.Close()
			Chain = kernel.NewChain(ChainPath, block)
			mempool = Chain.Mempool()
			Log().Info("SYNCABLE", i, block.Hash(), mempool.Height(), kernel.TXsSize)
		}

		if i == initHeight {
			continue
		}

		if i != 0 {
			ok := Chain.Accept(block)
			if !ok {
				Log().Error("SYNCABLE", i, []byte{0}, mempool.Height(), kernel.TXsSize)
				break
			}
			Log().Info("SYNCABLE", i, block.Hash(), mempool.Height(), kernel.TXsSize)
		}
	}
}

func handleGetTime(node network.Node, conn network.Conn, data []byte) {
	var (
		currTime = atomic.LoadUint64(&CurrentTime)
	)

	msg := network.NewMessage(
		MsgGetTime|MaskBit,
		encoding.Uint64ToBytes(currTime),
	)

	conn.Write(msg.Bytes())
}

func handleGetHeight(node network.Node, conn network.Conn, data []byte) {
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
	var (
		upBlock updateBlock
		mempool = Chain.Mempool()
		height  = Chain.Height()
	)

	err := json.Unmarshal(data, &upBlock)
	if err != nil {
		return
	}

	newBlock := kernel.LoadBlock(upBlock.Block)
	if newBlock == nil {
		return
	}

	if upBlock.Height < height {
		return
	}

	if upBlock.Height > height {
		if upBlock.Height > height+1 {
			Log().Error("MERGE", height, []byte{0}, mempool.Height(), kernel.TXsSize)
			return
		}
		for _, tx := range newBlock.Transactions() {
			if Chain.TX(tx.Hash()) != nil {
				continue
			}
			mempool.Push(tx)
		}
		tryUpdateBlock(node, mempool, height)
		return
	}

	if bytes.Equal(newBlock.Hash(), Chain.Block(height).Hash()) {
		return
	}

	// for {
	// 	ctime := atomic.LoadUint64(&CurrentTime) % IntervalTime
	// 	if ctime >= 1 && ctime <= IntervalTime-1 {
	// 		break
	// 	}
	// 	time.Sleep(200 * time.Millisecond)
	// }

	ok := Chain.Merge(height, newBlock.Transactions())
	if !ok {
		return
	}

	mergedBlock := Chain.Block(height)
	Log().Info("MERGE", height, mergedBlock.Hash(), mempool.Height(), kernel.TXsSize)

	upBlock = updateBlock{
		Height: height,
		Block:  mergedBlock.Bytes(),
	}

	upBlockBytes, err := json.Marshal(upBlock)
	if err != nil {
		return
	}

	node.Broadcast(network.NewMessage(MsgSetBlock, upBlockBytes))
}

func handleGetTX(node network.Node, conn network.Conn, data []byte) {
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
	var (
		mempool = Chain.Mempool()
		tx      = kernel.LoadTransaction(data)
		hash    = tx.Hash()
		retCode = uint64(0)
	)

	defer func(conn network.Conn) {
		msg := network.NewMessage(
			MsgSetTX|MaskBit,
			encoding.Uint64ToBytes(retCode),
		)
		conn.Write(msg.Bytes())
	}(conn)

	if tx == nil {
		return
	}

	txInChain := Chain.TX(hash)
	if txInChain != nil {
		retCode = 1
		return
	}

	txInMempool := mempool.TX(hash)
	if txInMempool != nil {
		retCode = 2
		return
	}

	mempool.Push(tx)
	node.Broadcast(network.NewMessage(MsgSetTX, data))
}

func tryUpdateBlock(node network.Node, mempool kernel.Mempool, height kernel.Height) {
	if mempool.Height() < kernel.TXsSize {
		return
	}

	txs := mempool.Pop()
	if txs == nil {
		return
	}

	lastBlock := Chain.Block(height)

	newBlock := kernel.NewBlock(lastBlock.Hash(), txs)
	newHeight := height + 1

	ok := Chain.Accept(newBlock)
	if !ok {
		Log().Error("ACCEPT", newHeight, []byte{0}, mempool.Height(), kernel.TXsSize)
		return
	}

	Log().Info("ACCEPT", newHeight, newBlock.Hash(), mempool.Height(), kernel.TXsSize)

	upBlock := updateBlock{
		Height: newHeight,
		Block:  newBlock.Bytes(),
	}

	upBlockBytes, err := json.Marshal(upBlock)
	if err != nil {
		return
	}

	node.Broadcast(network.NewMessage(MsgSetBlock, upBlockBytes))
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
