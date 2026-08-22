package node

import (
	"fmt"
	"net"
)

type Node struct {
	ID   string
	Port int
}

func New(id string, port int) *Node {
	return &Node{
		ID:   id,
		Port: port,
	}
}

func (n *Node) Start() error {
	address := fmt.Sprintf(":%d", n.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Printf("Node %s is listening on port %d\n", n.ID, n.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		fmt.Printf("Node %s accepted a connection from %s\n", n.ID, conn.RemoteAddr())
		conn.Close()
	}
}
