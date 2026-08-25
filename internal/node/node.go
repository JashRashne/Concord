package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/jashrashne/concord/internal/peer"
	"github.com/jashrashne/concord/internal/protocol"
	"github.com/jashrashne/concord/internal/store"
)

type Node struct {
	ID    string
	Port  int
	Peers []peer.Peer
	Store *store.Store
}

func New(id string, port int) *Node {
	return &Node{
		ID:    id,
		Port:  port,
		Peers: make([]peer.Peer, 0),
		Store: store.New(),
	}
}

func (n *Node) AddPeer(p peer.Peer) {
	n.Peers = append(n.Peers, p)
}

func (n *Node) Peer(id string) (peer.Peer, bool) {
	for _, p := range n.Peers {
		if p.ID == id {
			return p, true
		}
	}

	return peer.Peer{}, false
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
	for _, p := range n.Peers {
		fmt.Printf("Known peer: %s at %s\n", p.ID, p.Address)
	}

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

		case protocol.MessageTypeSet:
			n.Store.Set(message.Key, message.Value)

			fmt.Printf(
				"Node %s stored key=%s value=%s\n",
				n.ID,
				message.Key,
				message.Value,
			)

			response := protocol.Message{
				Type: protocol.MessageTypeOK,
				From: n.ID,
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println("failed to send OK:", err)
				return
			}

		case protocol.MessageTypeGet:
			value, ok := n.Store.Get(message.Key)

			if !ok {
				response := protocol.Message{
					Type: protocol.MessageTypeNotFound,
					From: n.ID,
				}

				if err := writeMessage(conn, response); err != nil {
					fmt.Println("failed to send NOT_FOUND:", err)
					return
				}

				break
			}

			response := protocol.Message{
				Type:  protocol.MessageTypeValue,
				From:  n.ID,
				Value: value,
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println("failed to send VALUE:", err)
				return
			}

		case protocol.MessageTypeDelete:
			n.Store.Delete(message.Key)

			fmt.Printf(
				"Node %s deleted key=%s\n",
				n.ID,
				message.Key,
			)

			response := protocol.Message{
				Type: protocol.MessageTypeOK,
				From: n.ID,
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println("failed to send OK:", err)
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

func (n *Node) Ping(address string, timeout time.Duration) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}

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
			return fmt.Errorf("failed waiting for PONG from %s: %w", address, err)
		}

		return fmt.Errorf("peer %s closed connection without responding", address)
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

func (n *Node) Get(address string, key string, timeout time.Duration) (string, bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", false, err
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return "", false, err
	}

	request := protocol.Message{
		Type: protocol.MessageTypeGet,
		From: n.ID,
		Key:  key,
	}

	if err := writeMessage(conn, request); err != nil {
		return "", false, err
	}

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", false, fmt.Errorf("failed waiting for GET response: %w", err)
		}

		return "", false, fmt.Errorf("peer closed connection without responding")
	}

	var response protocol.Message

	if err := json.Unmarshal([]byte(scanner.Text()), &response); err != nil {
		return "", false, err
	}

	switch response.Type {
	case protocol.MessageTypeValue:
		return response.Value, true, nil

	case protocol.MessageTypeNotFound:
		return "", false, nil

	default:
		return "", false, fmt.Errorf(
			"unexpected GET response type: %s",
			response.Type,
		)
	}
}

func (n *Node) SendAndWaitForOK(address string, message protocol.Message, timeout time.Duration) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}

	if err := writeMessage(conn, message); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed waiting for response: %w", err)
		}

		return fmt.Errorf("peer closed connection without responding")
	}

	var response protocol.Message

	if err := json.Unmarshal([]byte(scanner.Text()), &response); err != nil {
		return err
	}

	if response.Type != protocol.MessageTypeOK {
		return fmt.Errorf(
			"expected OK response, received %s",
			response.Type,
		)
	}

	return nil
}
