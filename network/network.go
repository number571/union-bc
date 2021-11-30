package network

import (
	"bytes"
	"github.com/number571/gopeer/crypto"
	"github.com/number571/gopeer/local"
	"net"
	"sync"
)

type Node struct {
	connections map[string]net.Conn
	mutex       sync.Mutex
}

func (node *Node) CreateNode() *Node {
	return &Node{
		connections: make(map[string]net.Conn),
	}
}

func (node *Node) hasMaxConnSize() bool {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	return uint(len(node.connections)) > CONN_SIZE
}

func (node *Node) Listen(address string) error {
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

		id := crypto.RandString(SALT_SIZE)

		node.setConnection(id, conn)
		go node.handleConn(id)
	}

	return nil
}

func (node *Node) handleConn(id string) {
	conn := node.getConnection(id)

	defer func() {
		conn.Close()
		node.delConnection(id)
	}()

	for {
		msg := readMessage(conn)
		if msg == nil {
			continue
		}

		node.send(msg)

	}
}

func (node *Node) send(msg []byte) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	for _, cn := range node.connections {
		go cn.Write(msg)
	}

}

func readMessage(conn net.Conn) []byte {
	const (
		SizeUint64 = 8 // bytes
	)

	var (
		msg    []byte
		size   = uint(0)
		buflen = make([]byte, SizeUint64)
		buffer = make([]byte, BUFF_SIZE)
	)

	length, err := conn.Read(buflen)
	if err != nil {
		return nil
	}
	if length != SizeUint64 {
		return nil
	}

	mustLen := local.Package(buflen).BytesToSize()
	if mustLen > PACK_SIZE {
		return nil
	}

	for {
		length, err = conn.Read(buffer)
		if err != nil {
			return nil
		}

		size += uint(length)
		if size > mustLen {
			return nil
		}

		msg = bytes.Join(
			[][]byte{
				msg,
				buffer[:length],
			},
			[]byte{},
		)

		if size == mustLen {
			break
		}
	}

	return msg
}

func (node *Node) setConnection(id string, conn net.Conn) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.connections[id] = conn
}

func (node *Node) getConnection(id string) net.Conn {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	return node.connections[id]
}

func (node *Node) delConnection(id string) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	delete(node.connections, id)
}
