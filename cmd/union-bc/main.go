package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/number571/go-peer/crypto"
	"github.com/number571/go-peer/encoding"
	"github.com/number571/union-bc/kernel"
	"github.com/number571/union-bc/network"
)

const (
	IntervalTime = 1 // seconds
)

var (
	Chain       kernel.Chain
	CurrentTime uint64
	ChainPath   = "chain" + os.Args[1]
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
	initClient(node)
}

func initNode(node network.Node) {
	var (
		address  = os.Args[1]
		listAddr = []string{
			":7070",
			":8080",
			":9090",
		}
	)

	// Connects
	for _, addr := range listAddr {
		if addr == address {
			continue
		}
		node.Connect(addr)
	}

	// Listen port
	fmt.Println("Node is listening...")
	if address != "" {
		go node.Listen(address)
	}

	// Get current height from nodes
	for _, addr := range listAddr {
		if addr == address {
			continue
		}

		conn := getConn(addr)
		if conn == nil {
			continue
		}

		syncBlocks(conn)
		break
	}

	// Get current time from nodes
	for _, addr := range listAddr {
		if addr == address {
			continue
		}

		conn := getConn(addr)
		if conn == nil {
			continue
		}

		atomic.StoreUint64(&CurrentTime, getTime(conn))
		break
	}

	// Run timer
	go func(node network.Node) {
		for {
			time.Sleep(1 * time.Second)
			atomic.AddUint64(&CurrentTime, 1)

			if CurrentTime%IntervalTime == 0 {
				tryUpdateBlock(node, Chain.Mempool(), Chain.Height())
			}
		}
	}(node)
}

func getConn(addr string) network.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil
	}
	return conn
}

func getTime(conn network.Conn) uint64 {
	msg := network.NewMessage(
		MsgGetTime,
		nil,
	)

	msg = network.SendRequest(conn, msg)
	if msg == nil {
		return 0
	}

	return encoding.BytesToUint64(msg.Body())
}

func syncBlocks(conn network.Conn) {
	msg := network.NewMessage(
		MsgGetHeight,
		nil,
	)

	msg = network.SendRequest(conn, msg)
	if msg == nil {
		return
	}

	initHeight := Chain.Height()
	height := encoding.BytesToUint64(msg.Body())

	mempool := Chain.Mempool()

	// syncable blocks
	for i := initHeight; i <= kernel.Height(height); i++ {
		msg := network.NewMessage(
			MsgGetBlock,
			encoding.Uint64ToBytes(uint64(i)),
		)

		msg = network.SendRequest(conn, msg)
		if msg == nil {
			break
		}

		newBlock := kernel.LoadBlock(msg.Body())
		if newBlock == nil {
			break
		}

		if i == 0 {
			Chain.Close()
			Chain = kernel.NewChain(ChainPath, newBlock)
			mempool = Chain.Mempool()
		}

		if i != 0 {
			ok := Chain.Accept(newBlock)
			if !ok {
				log.Printf("[E] %-15sheight=%020d hash=%064d mempool=%05d\n", "FAIL(SYNCABLE)", i, 0, mempool.Height())
				break
			}
			log.Printf("[I] %-15sheight=%020d hash=%X mempool=%05d\n", "SYNCABLE", i, newBlock.Hash(), mempool.Height())
		}
	}
}

func initClient(node network.Node) {
	for {
		command := inputString("")
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
			for i := 0; i < 200; i++ {
				msg := []byte(crypto.RandString(20))
				tx := kernel.NewTransaction(crypto.NewPrivKey(kernel.KeySize), msg)
				node.Broadcast(network.NewMessage(MsgSetTX, tx.Bytes()))
			}

		case "net":
			fmt.Println(len(node.Connections()))

		case "mp":
			fmt.Println(Chain.Mempool().Height())

		case "tm":
			fmt.Println(atomic.LoadUint64(&CurrentTime))
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
		upBlock   updateBlock
		mempool   = Chain.Mempool()
		height    = Chain.Height()
		lastBlock = Chain.Block(height)
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
		for _, tx := range newBlock.Transactions() {
			mempool.Push(tx)
		}
		return
	}

	if upBlock.Height > height+1 {
		log.Printf("[E] %-15sheight=%020d hash=%064d mempool=%05d\n", "FAIL(MERGE)", upBlock.Height, 0, mempool.Height())
		return
	}

	if bytes.Equal(newBlock.Hash(), lastBlock.Hash()) {
		return
	}

	ok := Chain.Merge(height, newBlock.Transactions())
	if !ok {
		return
	}

	log.Printf("[I] %-15sheight=%020d hash=%X mempool=%05d\n", "MERGE", height, newBlock.Hash(), mempool.Height())

	upBlock = updateBlock{
		Height: height,
		Block:  newBlock.Bytes(),
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
	node.Broadcast(network.NewMessage(MsgSetTX, data))
}

func tryUpdateBlock(node network.Node, mempool kernel.Mempool, height kernel.Height) {
	txs := []kernel.Transaction{}

	if mempool.Height() < kernel.TXsSize {
		return
	}

	txs = mempool.Pop()
	if txs == nil {
		return
	}

	lastBlock := Chain.Block(height)

	newBlock := kernel.NewBlock(lastBlock.Hash(), txs)
	newHeight := height + 1

	ok := Chain.Accept(newBlock)
	if !ok {
		log.Printf("[E] %-15sheight=%020d hash=%064d mempool=%05d\n", "FAIL(ACCEPT)", newHeight, 0, mempool.Height())
		return
	}

	log.Printf("[I] %-15sheight=%020d hash=%X mempool=%05d\n", "ACCEPT", newHeight, newBlock.Hash(), mempool.Height())

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
