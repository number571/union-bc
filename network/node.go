package network

import (
	"net"
	"sync"
)

var (
	_ Node = &NodeT{}
)

// Basic structure for network use.
type NodeT struct {
	mainMtx  sync.Mutex
	routeMtx sync.Mutex

	moniker      string
	connections  map[Conn]bool
	handleRoutes map[MsgType]HandleFunc
}

// Create client by private key as identification.
func NewNode(moniker string) Node {
	return &NodeT{
		moniker:      moniker,
		connections:  make(map[Conn]bool),
		handleRoutes: make(map[MsgType]HandleFunc),
	}
}

func (node *NodeT) Moniker() string {
	return node.moniker
}

func (node *NodeT) Broadcast(msg Message) {
	msgBytes := msg.Bytes()
	for _, conn := range node.Connections() {
		go conn.Write(msgBytes)
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

		node.setConnection(conn)
		go node.handleConn(conn)
	}

	return nil
}

// Add function to mapping for route use.
func (node *NodeT) Handle(tmsg MsgType, handle HandleFunc) Node {
	node.setFunction(tmsg, handle)
	return node
}

func (node *NodeT) handleConn(conn Conn) {
	defer func() {
		node.delConnection(conn)
	}()

	counter := 0
	for {
		if counter == RetrySize {
			break
		}

		msg := ReadMessage(conn)
		if msg == nil {
			counter++
			continue
		}

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

	f(node, conn, msg.Body())
	return true
}

// Get list of connection addresses.
func (node *NodeT) Connections() []Conn {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	var list []Conn
	for conn := range node.connections {
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

	node.setConnection(conn)
	go node.handleConn(conn)

	return conn
}

func (node *NodeT) Disconnect(conn Conn) {
	node.delConnection(conn)
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

func (node *NodeT) setConnection(conn Conn) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	node.connections[conn] = true
}

func (node *NodeT) delConnection(conn Conn) {
	node.mainMtx.Lock()
	defer node.mainMtx.Unlock()

	delete(node.connections, conn)
	conn.Close()
}
