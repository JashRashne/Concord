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

	"sync"

	"github.com/jashrashne/concord/internal/command"
	"github.com/jashrashne/concord/internal/wal"

	"path/filepath"

	"github.com/jashrashne/concord/internal/raft"
)

type Node struct {
	ID    string
	Port  int
	Peers []peer.Peer
	Store *store.Store

	mutationMu sync.Mutex
	wal        *wal.Log
	raftState  *raft.State
}

func New(id string, port int) *Node {
	return &Node{
		ID:        id,
		Port:      port,
		Peers:     make([]peer.Peer, 0),
		Store:     store.New(),
		raftState: raft.New(),
	}
}

func NewPersistent(
	id string,
	port int,
	walPath string,
) (*Node, error) {
	log, err := wal.Open(walPath)
	if err != nil {
		return nil, err
	}

	raftPath := filepath.Join(
		filepath.Dir(walPath),
		"raft-state.json",
	)

	persistentRaftState, err := raft.Open(raftPath)
	if err != nil {
		log.Close()

		return nil, fmt.Errorf(
			"open persistent Raft state: %w",
			err,
		)
	}

	n := New(id, port)
	n.raftState = persistentRaftState
	n.wal = log

	records, err := log.ReadAll()
	if err != nil {
		log.Close()
		return nil, fmt.Errorf("recover WAL: %w", err)
	}

	for _, record := range records {
		if err := n.applyCommand(record.Command, false); err != nil {
			log.Close()
			return nil, fmt.Errorf("replay WAL: %w", err)
		}
	}

	return n, nil
}

func (n *Node) applyCommand(
	cmd command.Command,
	persist bool,
) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	n.mutationMu.Lock()
	defer n.mutationMu.Unlock()

	if persist && n.wal != nil {
		if err := n.wal.Append(wal.NewRecord(cmd)); err != nil {
			return fmt.Errorf("append command to WAL: %w", err)
		}
	}

	switch cmd.Type {
	case command.TypeSet:
		n.Store.Set(cmd.Key, cmd.Value)

	case command.TypeDelete:
		n.Store.Delete(cmd.Key)

	default:
		return fmt.Errorf(
			"unsupported command type %q",
			cmd.Type,
		)
	}

	return nil
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
			cmd := command.Command{
				Type:  command.TypeSet,
				Key:   message.Key,
				Value: message.Value,
			}

			if err := n.applyCommand(cmd, true); err != nil {
				fmt.Println("failed to apply SET:", err)
				return
			}

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
			cmd := command.Command{
				Type: command.TypeDelete,
				Key:  message.Key,
			}

			if err := n.applyCommand(cmd, true); err != nil {
				fmt.Println("failed to apply DELETE:", err)
				return
			}

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

		case protocol.MessageTypeRequestVote:
			if message.RequestVote == nil {
				fmt.Println("REQUEST_VOTE missing payload")
				return
			}

			request := message.RequestVote

			term, granted, err := n.raftState.HandleRequestVote(
				request.Term,
				request.CandidateID,
			)
			if err != nil {
				fmt.Println(
					"failed to process REQUEST_VOTE:",
					err,
				)
				return
			}

			response := protocol.Message{
				Type: protocol.MessageTypeRequestVoteResponse,
				From: n.ID,
				RequestVoteResponse: &protocol.RequestVoteResponse{
					Term:        term,
					VoteGranted: granted,
				},
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println(
					"failed to send REQUEST_VOTE_RESPONSE:",
					err,
				)
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

func (n *Node) sendRequest(
	address string,
	request protocol.Message,
	timeout time.Duration,
) (protocol.Message, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return protocol.Message{}, err
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return protocol.Message{}, err
	}

	if err := writeMessage(conn, request); err != nil {
		return protocol.Message{}, err
	}

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return protocol.Message{}, fmt.Errorf(
				"failed waiting for response from %s: %w",
				address,
				err,
			)
		}

		return protocol.Message{}, fmt.Errorf(
			"peer %s closed connection without responding",
			address,
		)
	}

	var response protocol.Message

	if err := json.Unmarshal([]byte(scanner.Text()), &response); err != nil {
		return protocol.Message{}, err
	}

	return response, nil
}

func (n *Node) Ping(address string, timeout time.Duration) error {
	request := protocol.Message{
		Type: protocol.MessageTypePing,
		From: n.ID,
	}

	response, err := n.sendRequest(address, request, timeout)
	if err != nil {
		return err
	}

	if response.Type != protocol.MessageTypePong {
		return fmt.Errorf(
			"expected PONG, received %s",
			response.Type,
		)
	}

	fmt.Printf(
		"Node %s received PONG from %s\n",
		n.ID,
		response.From,
	)

	return nil
}

func (n *Node) Get(
	address string,
	key string,
	timeout time.Duration,
) (string, bool, error) {
	request := protocol.Message{
		Type: protocol.MessageTypeGet,
		From: n.ID,
		Key:  key,
	}

	response, err := n.sendRequest(address, request, timeout)
	if err != nil {
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

func (n *Node) SendAndWaitForOK(
	address string,
	message protocol.Message,
	timeout time.Duration,
) error {
	response, err := n.sendRequest(address, message, timeout)
	if err != nil {
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

func (n *Node) Close() error {
	if n.wal == nil {
		return nil
	}

	return n.wal.Close()
}
func (n *Node) RaftSnapshot() raft.Snapshot {
	return n.raftState.Snapshot()
}

func (n *Node) RequestVote(
	address string,
	request protocol.RequestVoteRequest,
	timeout time.Duration,
) (protocol.RequestVoteResponse, error) {
	message := protocol.Message{
		Type:        protocol.MessageTypeRequestVote,
		From:        n.ID,
		RequestVote: &request,
	}

	response, err := n.sendRequest(
		address,
		message,
		timeout,
	)
	if err != nil {
		return protocol.RequestVoteResponse{}, err
	}

	if response.Type != protocol.MessageTypeRequestVoteResponse {
		return protocol.RequestVoteResponse{},
			fmt.Errorf(
				"expected REQUEST_VOTE_RESPONSE, received %s",
				response.Type,
			)
	}

	if response.RequestVoteResponse == nil {
		return protocol.RequestVoteResponse{},
			fmt.Errorf(
				"REQUEST_VOTE_RESPONSE missing payload",
			)
	}

	return *response.RequestVoteResponse, nil
}
