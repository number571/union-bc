package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
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

	if len(os.Args) >= 3 && os.Args[2] == "rollback" {
		defaultNum := 10
		if len(os.Args) == 4 {
			defaultNum, _ = strconv.Atoi(os.Args[3])
		}
		Chain.Rollback(uint64(defaultNum))
		os.Exit(1)
	}
}

func main() {
	node := network.NewNode().
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

	for i := 0; i < ClientsNum; i++ {
		go func() {
			conn := network.NewConn(Address)
			if conn == nil {
				panic("conn is nil")
			}
			defer conn.Close()

			for {
				priv := crypto.NewPrivKey(kernel.KeySize)
				for i := 0; i < TXsInSecond; i++ {
					tx := kernel.NewTransaction(priv, []byte(crypto.RandString(20)))
					_ = conn.Request(network.NewMessage(MsgSetTX, tx.Bytes()))
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	select {}
}

func initNode(node network.Node) {
	fmt.Println("Node is listening...")
	var conn network.Conn

	for _, addr := range ListAddr {
		if addr == Address {
			continue
		}
		conn = network.NewConn(addr)
		if conn == nil {
			continue
		}
		break
	}

	if conn != nil {
		syncBlocks(conn)
		atomic.StoreUint64(&CurrentTime, getTime(conn))
		conn.Close()
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

	// Generate block
	go func(node network.Node) {
		for {
			time.Sleep(1 * time.Second)
			atomic.AddUint64(&CurrentTime, 1)

			ctime := atomic.LoadUint64(&CurrentTime) % IntervalTime
			if ctime != 0 {
				continue
			}

			commitBlock(node, Chain.Mempool(), Chain.Height())
			tryUpdateBlock(node, Chain.Mempool(), Chain.Height())
		}
	}(node)
}

func commitBlock(node network.Node, mempool kernel.Mempool, height kernel.Height) {
	type blockInfo struct {
		count uint
		block kernel.Block
	}

	var (
		commitBlock = Chain.Block(height)
		hash        = encoding.Base64Encode(commitBlock.Hash())
		blocks      = make(map[string]blockInfo)
	)

	blocks[hash] = blockInfo{
		count: 1,
		block: commitBlock,
	}

	for _, addr := range ListAddr {
		if addr == Address {
			continue
		}

		conn := network.NewConn(addr)
		if conn == nil {
			continue
		}

		block := getBlock(conn, height)
		conn.Close()
		if block == nil {
			continue
		}

		if !bytes.Equal(block.PrevHash(), commitBlock.PrevHash()) {
			continue
		}

		hash := encoding.Base64Encode(block.Hash())
		if val, ok := blocks[hash]; !ok {
			blocks[hash] = blockInfo{
				count: 1,
				block: block,
			}
		} else {
			blocks[hash] = blockInfo{
				count: val.count + 1,
				block: block,
			}
		}
	}

	var listBlocks []blockInfo
	for _, val := range blocks {
		listBlocks = append(listBlocks, val)
	}

	sort.SliceStable(listBlocks, func(i, j int) bool {
		return listBlocks[i].count > listBlocks[j].count
	})

	maxCount := listBlocks[0].count
	for i, info := range listBlocks {
		if info.count < maxCount {
			listBlocks = listBlocks[:i]
			break
		}
	}

	sort.SliceStable(listBlocks, func(i, j int) bool {
		return bytes.Compare(listBlocks[i].block.Hash(), listBlocks[j].block.Hash()) < 0
	})

	if bytes.Equal(commitBlock.Hash(), listBlocks[0].block.Hash()) {
		Log().Info("COMMIT", height, commitBlock.Hash(), mempool.Height(), kernel.TXsSize, len(node.Connections()))
		return
	}

	commitBlock = listBlocks[0].block

	node.Mutex().Lock()
	ok := Chain.Rollback(1)
	node.Mutex().Unlock()
	if !ok {
		Log().Warning("COMMIT", height, commitBlock.Hash(), mempool.Height(), kernel.TXsSize, len(node.Connections()))
		return
	}

	ok = Chain.Accept(commitBlock)
	if !ok {
		Log().Error("COMMIT", height, mempool.Height(), kernel.TXsSize, len(node.Connections()))
		return
	}

	Log().Info("COMMIT", height, commitBlock.Hash(), mempool.Height(), kernel.TXsSize, len(node.Connections()))
}

func getBlock(conn network.Conn, height kernel.Height) kernel.Block {
	msg := network.NewMessage(
		MsgGetBlock,
		encoding.Uint64ToBytes(uint64(height)),
	)

	msg = conn.Request(msg)
	if msg == nil {
		return nil
	}

	return kernel.LoadBlock(msg.Body())
}

func getTime(conn network.Conn) uint64 {
	msg := network.NewMessage(
		MsgGetTime,
		nil,
	)

	msg = conn.Request(msg)
	if msg == nil {
		return 0
	}

	return encoding.BytesToUint64(msg.Body())
}

func syncBlocks(conn network.Conn) {
	var (
		mempool    = Chain.Mempool()
		initHeight = Chain.Height()
	)

	i := initHeight

	// syncable all blocks
	for ; ; i++ {
		if i == initHeight && initHeight != 0 {
			continue
		}

		block := getBlock(conn, i)
		if block == nil {
			break
		}

		if i == 0 {
			Chain.Close()
			Chain = kernel.NewChain(ChainPath, block)
			mempool = Chain.Mempool()
			Log().Warning("SYNCABLE", i, block.Hash(), mempool.Height(), kernel.TXsSize, 0)
		}

		if i != 0 {
			ok := Chain.Accept(block)
			if !ok {
				Log().Error("SYNCABLE", i, mempool.Height(), kernel.TXsSize, 0)
				os.Exit(1)
			}
			Log().Info("SYNCABLE", i, block.Hash(), mempool.Height(), kernel.TXsSize, 0)
		}
	}
}

func handleGetTime(node network.Node, conn network.Conn, msg network.Message) {
	var (
		currTime = atomic.LoadUint64(&CurrentTime)
	)

	rmsg := network.NewMessage(
		MsgGetTime|MaskBit,
		encoding.Uint64ToBytes(currTime),
	)

	conn.Write(rmsg)
}

func handleGetHeight(node network.Node, conn network.Conn, msg network.Message) {
	var (
		height = uint64(Chain.Height())
	)

	rmsg := network.NewMessage(
		MsgGetHeight|MaskBit,
		encoding.Uint64ToBytes(height),
	)

	conn.Write(rmsg)
}

func handleGetBlock(node network.Node, conn network.Conn, msg network.Message) {
	var (
		height     = kernel.Height(encoding.BytesToUint64(msg.Body()))
		block      = Chain.Block(height)
		blockBytes = []byte{}
	)

	if block != nil {
		blockBytes = block.Bytes()
	}

	rmsg := network.NewMessage(
		MsgGetBlock|MaskBit,
		blockBytes,
	)

	conn.Write(rmsg)
}

func handleSetBlock(node network.Node, conn network.Conn, msg network.Message) {
	var (
		mempool   = Chain.Mempool()
		height    = Chain.Height()
		currBlock = Chain.Block(height)
	)

	upBlock := updateBlock{}
	err := json.Unmarshal(msg.Body(), &upBlock)
	if err != nil {
		return
	}

	newBlock := kernel.LoadBlock(upBlock.Block)
	if newBlock == nil {
		return
	}

	if upBlock.Height != height {
		for _, tx := range newBlock.Transactions() {
			if Chain.TX(tx.Hash()) != nil {
				continue
			}
			mempool.Push(tx)
		}

		node.Broadcast(msg)
		return
	}

	if bytes.Equal(newBlock.Hash(), currBlock.Hash()) {
		return
	}

	ok := Chain.Merge(height, newBlock.Transactions())
	if !ok {
		return
	}

	mergedBlock := Chain.Block(height)
	Log().Info("MERGE", height, mergedBlock.Hash(), mempool.Height(), kernel.TXsSize, len(node.Connections()))

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

func handleGetTX(node network.Node, conn network.Conn, msg network.Message) {
	var (
		hash    = kernel.Hash(msg.Body())
		tx      = Chain.TX(hash)
		txBytes = []byte{}
	)

	if tx != nil {
		txBytes = tx.Bytes()
	}

	rmsg := network.NewMessage(
		MsgGetTX|MaskBit,
		txBytes,
	)

	conn.Write(rmsg)
}

func handleSetTX(node network.Node, conn network.Conn, msg network.Message) {
	var (
		mempool = Chain.Mempool()
		tx      = kernel.LoadTransaction(msg.Body())
		hash    = tx.Hash()
		retCode = uint64(0)
	)

	defer func(conn network.Conn) {
		msg := network.NewMessage(
			MsgSetTX|MaskBit,
			encoding.Uint64ToBytes(retCode),
		)
		conn.Write(msg)
	}(conn)

	if tx == nil {
		retCode = 2
		return
	}

	txInChain := Chain.TX(hash)
	if txInChain != nil {
		retCode = 3
		return
	}

	txInMempool := mempool.TX(hash)
	if txInMempool != nil {
		retCode = 4
		return
	}

	mempool.Push(tx)
}

func tryUpdateBlock(node network.Node, mempool kernel.Mempool, height kernel.Height) {
	node.Mutex().Lock()
	defer node.Mutex().Unlock()

	txs := mempool.Pop()
	if txs == nil {
		return
	}

	lastBlock := Chain.Block(height)

	newBlock := kernel.NewBlock(lastBlock.Hash(), txs)
	newHeight := height + 1

	ok := Chain.Accept(newBlock)
	if !ok {
		Log().Error("ACCEPT", newHeight, mempool.Height(), kernel.TXsSize, len(node.Connections()))
		return
	}

	Log().Info("ACCEPT", newHeight, newBlock.Hash(), mempool.Height(), kernel.TXsSize, len(node.Connections()))

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
