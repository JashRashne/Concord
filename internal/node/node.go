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

func (n *Node) Send(address string, message string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		return err
	}

	fmt.Printf("Node %s sent message to %s: %s\n", n.ID, address, message)

	return nil
}

func (n *Node) Start() error {
	address := fmt.Sprintf(":%d", n.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Node %s listening on port %d\n", n.ID, n.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("Node %s accepted connection from %s\n", n.ID, conn.RemoteAddr())

	buffer := make([]byte, 1024)

	bytesRead, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("failed to read from connection:", err)
		return
	}

	message := string(buffer[:bytesRead])

	fmt.Printf("Node %s received: %s\n", n.ID, message)
}
