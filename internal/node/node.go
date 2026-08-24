package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/jashrashne/concord/internal/protocol"
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

func (n *Node) Send(address string, message protocol.Message) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := writeMessage(conn, message); err != nil {
		return err
	}

	fmt.Printf("Node %s sent message to %s\n", n.ID, address)

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

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		var message protocol.Message

		if err := json.Unmarshal([]byte(scanner.Text()), &message); err != nil {
			fmt.Println("failed to decode message:", err)
			continue
		}

		switch message.Type {
		case protocol.MessageTypePing:
			fmt.Printf("Node %s received PING from %s\n", n.ID, message.From)

			response := protocol.Message{
				Type: protocol.MessageTypePong,
				From: n.ID,
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println("failed to send PONG:", err)
				return
			}

		default:
			fmt.Printf(
				"Node %s received message from %s: type=%s data=%s\n",
				n.ID,
				message.From,
				message.Type,
				message.Data,
			)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("failed to read from connection:", err)
	}
}

func writeMessage(conn net.Conn, message protocol.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = conn.Write(data)
	return err
}

func (n *Node) Ping(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	ping := protocol.Message{
		Type: protocol.MessageTypePing,
		From: n.ID,
	}

	if err := writeMessage(conn, ping); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}

		return fmt.Errorf("peer closed connection without responding")
	}

	var response protocol.Message

	if err := json.Unmarshal([]byte(scanner.Text()), &response); err != nil {
		return err
	}

	if response.Type != protocol.MessageTypePong {
		return fmt.Errorf("expected PONG, received %s", response.Type)
	}

	fmt.Printf("Node %s received PONG from %s\n", n.ID, response.From)

	return nil
}
