package main

import (
	"fmt"
	"time"

	"github.com/number571/unionbc/network"
)

const (
	SEED_NODE_ADDR = ":8080"
)

const (
	MsgEcho network.MsgType = 1 + iota
)

func main() {
	nodes := []network.Node{}

	for i := 0; i < 3; i++ {
		node := network.NewNode(fmt.Sprintf("moniker-%d", i)).
			Handle(MsgEcho, handleEcho)
		nodes = append(nodes, node)
	}

	go nodes[0].Listen(SEED_NODE_ADDR)
	time.Sleep(500 * time.Millisecond)

	for i := 1; i < 3; i++ {
		nodes[i].Connect(SEED_NODE_ADDR)
	}

	fmt.Println("Message sent")

	msg := network.NewMessage(MsgEcho,
		[]byte("hello, world!"))
	nodes[2].Broadcast(msg)

	select {}
}

func handleEcho(node network.Node, conn network.Conn, data []byte) {
	fmt.Println(node.Moniker(), string(data))

	// msg := network.NewMessage(MsgEcho, []byte(data))
	// node.Broadcast(msg)
}
