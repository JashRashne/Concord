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

	electionResetCh chan struct{}

	stopCh    chan struct{}
	closeOnce sync.Once
	closeErr  error

	listenerMu sync.Mutex
	listener   net.Listener

	raftWG sync.WaitGroup
}

type voteResult struct {
	response protocol.RequestVoteResponse
	err      error
}

func New(id string, port int) *Node {
	return &Node{
		ID:              id,
		Port:            port,
		Peers:           make([]peer.Peer, 0),
		Store:           store.New(),
		raftState:       raft.New(),
		electionResetCh: make(chan struct{}, 1),
		stopCh:          make(chan struct{}),
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

	select {
	case <-n.stopCh:
		return fmt.Errorf("node is already closed")

	default:
	}

	address := fmt.Sprintf(":%d", n.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	n.listenerMu.Lock()
	n.listener = listener
	n.listenerMu.Unlock()

	defer listener.Close()

	n.raftWG.Add(2)

	go func() {
		defer n.raftWG.Done()
		n.runElectionLoop()
	}()

	go func() {
		defer n.raftWG.Done()
		n.runHeartbeatLoop()
	}()

	fmt.Printf("Node %s listening on port %d\n", n.ID, n.Port)
	for _, p := range n.Peers {
		fmt.Printf("Known peer: %s at %s\n", p.ID, p.Address)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-n.stopCh:
				return nil

			default:
				return err
			}
		}

		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	// removed heartbeat/connecting msg because heartbeat is generating a lot of msg noise
	// fmt.Printf("Node %s accepted connection from %s\n", n.ID, conn.RemoteAddr())

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

			term, granted, err :=
				n.raftState.HandleRequestVote(
					request.Term,
					request.CandidateID,
					request.LastLogIndex,
					request.LastLogTerm,
				)
			if err != nil {
				fmt.Println(
					"failed to process REQUEST_VOTE:",
					err,
				)
				return
			}
			if granted {
				n.resetElectionTimer()
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

		case protocol.MessageTypeAppendEntries:
			if message.AppendEntries == nil {
				fmt.Println("APPEND_ENTRIES missing payload")
				return
			}

			before := n.raftState.Snapshot()

			request := message.AppendEntries

			term, success, matchIndex, err :=
				n.raftState.HandleAppendEntries(
					request.Term,
					request.LeaderID,
					request.PrevLogIndex,
					request.PrevLogTerm,
					request.Entries,
				)

			if err != nil {
				fmt.Println(
					"failed to process APPEND_ENTRIES:",
					err,
				)
				return
			}

			if request.Term >= before.CurrentTerm {
				n.resetElectionTimer()
			}

			after := n.raftState.Snapshot()

			if success &&
				(before.LeaderID != after.LeaderID ||
					before.CurrentTerm != after.CurrentTerm ||
					before.Role != after.Role) {

				fmt.Printf(
					"Node %s following leader %s in term %d\n",
					n.ID,
					after.LeaderID,
					after.CurrentTerm,
				)
			}

			response := protocol.Message{
				Type: protocol.MessageTypeAppendEntriesResponse,
				From: n.ID,
				AppendEntriesResponse: &protocol.AppendEntriesResponse{
					Term:       term,
					Success:    success,
					MatchIndex: matchIndex,
				},
			}

			if err := writeMessage(conn, response); err != nil {
				fmt.Println(
					"failed to send APPEND_ENTRIES_RESPONSE:",
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
	n.closeOnce.Do(func() {
		close(n.stopCh)

		n.listenerMu.Lock()

		if n.listener != nil {
			_ = n.listener.Close()
		}

		n.listenerMu.Unlock()

		n.raftWG.Wait()

		if n.wal != nil {
			n.closeErr = n.wal.Close()
		}
	})

	return n.closeErr
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

func (n *Node) resetElectionTimer() {
	select {
	case n.electionResetCh <- struct{}{}:
	default:
	}
}

func (n *Node) runElectionLoop() {
	timer := time.NewTimer(
		raft.RandomElectionTimeout(),
	)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			snapshot := n.raftState.Snapshot()

			if snapshot.Role != raft.RoleLeader {
				n.startElection()
			}

			timer.Reset(
				raft.RandomElectionTimeout(),
			)

		case <-n.electionResetCh:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			timer.Reset(
				raft.RandomElectionTimeout(),
			)

		case <-n.stopCh:
			return
		}
	}
}

func (n *Node) startElection() {
	term, err := n.raftState.StartElection(n.ID)
	if err != nil {
		fmt.Printf(
			"Node %s failed to start election: %v\n",
			n.ID,
			err,
		)
		return
	}

	fmt.Printf(
		"Node %s started election for term %d\n",
		n.ID,
		term,
	)

	votes := 1

	clusterSize := len(n.Peers) + 1
	majority := raft.Majority(clusterSize)

	if votes >= majority {
		if n.raftState.BecomeLeader(term, n.ID) {
			fmt.Printf(
				"Node %s became LEADER for term %d with %d votes\n",
				n.ID,
				term,
				votes,
			)

			n.sendHeartbeatRound(term)
		}

		return
	}

	results := make(
		chan voteResult,
		len(n.Peers),
	)

	lastLogIndex, lastLogTerm :=
		n.raftState.LastLogInfo()

	request := protocol.RequestVoteRequest{
		Term:         term,
		CandidateID:  n.ID,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	for _, p := range n.Peers {
		p := p

		go func() {
			response, err := n.RequestVote(
				p.Address,
				request,
				raft.RequestVoteTimeout,
			)

			results <- voteResult{
				response: response,
				err:      err,
			}
		}()
	}

	for range n.Peers {
		result := <-results

		if result.err != nil {
			continue
		}

		if result.response.Term > term {
			_, err := n.raftState.ObserveTerm(
				result.response.Term,
			)

			if err != nil {
				fmt.Printf(
					"Node %s failed to observe higher term: %v\n",
					n.ID,
					err,
				)
			}

			return
		}

		snapshot := n.raftState.Snapshot()

		if snapshot.Role != raft.RoleCandidate ||
			snapshot.CurrentTerm != term {
			return
		}

		if result.response.Term != term {
			continue
		}

		if !result.response.VoteGranted {
			continue
		}

		votes++

		if votes >= majority {
			if n.raftState.BecomeLeader(term, n.ID) {
				fmt.Printf(
					"Node %s became LEADER for term %d with %d votes\n",
					n.ID,
					term,
					votes,
				)

				n.sendHeartbeatRound(term)
			}

			return
		}
	}
}

func (n *Node) AppendEntries(
	address string,
	request protocol.AppendEntriesRequest,
	timeout time.Duration,
) (protocol.AppendEntriesResponse, error) {
	message := protocol.Message{
		Type:          protocol.MessageTypeAppendEntries,
		From:          n.ID,
		AppendEntries: &request,
	}

	response, err := n.sendRequest(
		address,
		message,
		timeout,
	)
	if err != nil {
		return protocol.AppendEntriesResponse{}, err
	}

	if response.Type != protocol.MessageTypeAppendEntriesResponse {
		return protocol.AppendEntriesResponse{},
			fmt.Errorf(
				"expected APPEND_ENTRIES_RESPONSE, received %s",
				response.Type,
			)
	}

	if response.AppendEntriesResponse == nil {
		return protocol.AppendEntriesResponse{},
			fmt.Errorf(
				"APPEND_ENTRIES_RESPONSE missing payload",
			)
	}

	return *response.AppendEntriesResponse, nil
}

func (n *Node) sendHeartbeatRound(term uint64) {
	var wg sync.WaitGroup

	request := protocol.AppendEntriesRequest{
		Term:         term,
		LeaderID:     n.ID,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      nil,
		LeaderCommit: 0,
	}

	for _, p := range n.Peers {
		p := p

		wg.Add(1)

		go func() {
			defer wg.Done()

			response, err := n.AppendEntries(
				p.Address,
				request,
				raft.HeartbeatRPCTimeout,
			)
			if err != nil {
				return
			}

			if response.Term <= term {
				return
			}

			changed, err := n.raftState.ObserveTerm(
				response.Term,
			)
			if err != nil {
				fmt.Printf(
					"Node %s failed to observe higher term %d: %v\n",
					n.ID,
					response.Term,
					err,
				)

				return
			}

			if changed {
				fmt.Printf(
					"Node %s stepped down after discovering term %d\n",
					n.ID,
					response.Term,
				)

				n.resetElectionTimer()
			}
		}()
	}

	wg.Wait()
}

func (n *Node) runHeartbeatLoop() {
	ticker := time.NewTicker(
		raft.HeartbeatInterval,
	)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			snapshot := n.raftState.Snapshot()

			if snapshot.Role != raft.RoleLeader {
				continue
			}

			n.sendHeartbeatRound(
				snapshot.CurrentTerm,
			)

		case <-n.stopCh:
			return
		}
	}
}
