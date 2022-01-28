package network

import (
	"net"
	"sync"

	"github.com/number571/go-peer/crypto"
)

var (
	_ Node = &NodeT{}
)

// Basic structure for network use.
type NodeT struct {
	mainMtx  sync.Mutex
	routeMtx sync.Mutex

	mapping      map[string]bool
	connections  map[string]Conn
	handleRoutes map[MsgType]HandleFunc
}

// Create client by private key as identification.
func NewNode() Node {
	return &NodeT{
		mapping:      make(map[string]bool),
		connections:  make(map[string]Conn),
		handleRoutes: make(map[MsgType]HandleFunc),
	}
}

func (node *NodeT) Mutex() *sync.Mutex {
	return &node.routeMtx
}

func (node *NodeT) Broadcast(msg Message) {
	node.setMapping(msg.Hash())

	for _, conn := range node.Connections() {
		go conn.Write(msg)
	}
}

// Turn on listener by address.
// Client handle function need be not null.
func (node *NodeT) Listen(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			break
		}

		if node.hasMaxConnSize() {
			conn.Close()
			continue
		}

		var (
			iconn = &ConnT{crypto.RandString(16), conn}
			whoIs = make([]byte, 1)
		)
		conn.Read(whoIs)

		switch whoIs[0] {
		case IsNode:
			node.setConnection(iconn)
		case IsClient:
			// do nothing
		default:
			conn.Close()
			continue
		}

		go node.handleConn(iconn)
	}

	return nil
}

// Add function to mapping for route use.
func (node *NodeT) Handle(tmsg MsgType, handle HandleFunc) Node {
	node.setFunction(tmsg, handle)
	return node
}

func (node *NodeT) handleConn(conn *ConnT) {
	defer func() {
		node.delConnection(conn)
	}()

	counter := 0

	for counter != RetrySize {
		msg := conn.Read()
		if msg == nil {
			counter++
			continue
		}

		hash := msg.Hash()
		if node.inMapping(hash) {
			counter++
			continue
		}
		node.setMapping(hash)

		ok := node.handleFunc(conn, msg)
		if !ok {
			counter++
			continue
		}

		counter = 0
	}
}

func (node *NodeT) handleFunc(conn Conn, msg Message) bool {
	node.routeMtx.Lock()
	defer node.routeMtx.Unlock()

	f, ok := node.getFunction(msg.Head())
	if !ok {
		return false
	}

	f(node, conn, msg)
	return true
}

// Get list of connection addresses.
func (node *NodeT) Connections() []Conn {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	var list []Conn
	for _, conn := range node.connections {
		list = append(list, conn)
	}

	return list
}

// Connect to node by address.
// Client handle function need be not null.
func (node *NodeT) Connect(address string) Conn {
	if node.hasMaxConnSize() {
		return nil
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}

	conn.Write([]byte{IsNode})

	iconn := &ConnT{crypto.RandString(16), conn}

	node.setConnection(iconn)
	go node.handleConn(iconn)

	return iconn
}

func (node *NodeT) Disconnect(conn Conn) {
	node.delConnection(conn.(*ConnT))
}

func (node *NodeT) setFunction(tmsg MsgType, handle HandleFunc) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	node.handleRoutes[tmsg] = handle
}

func (node *NodeT) getFunction(tmsg MsgType) (HandleFunc, bool) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	f, ok := node.handleRoutes[tmsg]
	return f, ok
}

func (node *NodeT) hasMaxConnSize() bool {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	return len(node.connections) > ConnSize
}

func (node *NodeT) setConnection(conn *ConnT) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	node.connections[conn.nonce] = conn
}

func (node *NodeT) delConnection(conn *ConnT) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	delete(node.connections, conn.nonce)
	conn.Close()
}

func (node *NodeT) inMapping(hash string) bool {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	_, ok := node.mapping[hash]
	return ok
}

func (node *NodeT) setMapping(hash string) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	if uint(len(node.mapping)) > MappSize {
		for k := range node.mapping {
			delete(node.mapping, k)
			break
		}
	}

	node.mapping[hash] = true
}
