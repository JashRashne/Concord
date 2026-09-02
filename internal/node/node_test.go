package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/jashrashne/concord/internal/command"
	"github.com/jashrashne/concord/internal/peer"
	"github.com/jashrashne/concord/internal/protocol"

	"path/filepath"
)

func TestPeerFindsPeerByID(t *testing.T) {
	n := New("node-1", 8080)

	n.AddPeer(peer.Peer{
		ID:      "node-2",
		Address: "localhost:8081",
	})

	n.AddPeer(peer.Peer{
		ID:      "node-3",
		Address: "localhost:8082",
	})

	got, ok := n.Peer("node-3")

	if !ok {
		t.Fatal("expected to find node-3")
	}

	if got.ID != "node-3" {
		t.Fatalf("expected peer ID node-3, got %s", got.ID)
	}

	if got.Address != "localhost:8082" {
		t.Fatalf(
			"expected address localhost:8082, got %s",
			got.Address,
		)
	}
}

func TestPeerReturnsFalseForUnknownPeer(t *testing.T) {
	n := New("node-1", 8080)

	n.AddPeer(peer.Peer{
		ID:      "node-2",
		Address: "localhost:8081",
	})

	got, ok := n.Peer("node-99")

	if ok {
		t.Fatal("expected unknown peer lookup to return false")
	}

	if got != (peer.Peer{}) {
		t.Fatalf("expected zero-value peer, got %+v", got)
	}
}

func TestNewNodeCreatesStore(t *testing.T) {
	n := New("node-1", 8080)

	if n.Store == nil {
		t.Fatal("expected node store to be initialized")
	}
}

func TestNodesHaveIndependentStores(t *testing.T) {
	node1 := New("node-1", 8080)
	node2 := New("node-2", 8081)

	node1.Store.Set("name", "alice")

	_, ok := node2.Store.Get("name")

	if ok {
		t.Fatal("expected node stores to be independent")
	}
}

func TestSendAndWaitForOKSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				serverErr <- err
				return
			}

			serverErr <- fmt.Errorf("client closed connection without sending a message")
			return
		}

		var request protocol.Message

		if err := json.Unmarshal([]byte(scanner.Text()), &request); err != nil {
			serverErr <- err
			return
		}

		response := protocol.Message{
			Type: protocol.MessageTypeOK,
			From: "server",
		}

		if err := writeMessage(conn, response); err != nil {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	n := New("client", 0)

	message := protocol.Message{
		Type:  protocol.MessageTypeSet,
		From:  n.ID,
		Key:   "name",
		Value: "alice",
	}

	err = n.SendAndWaitForOK(
		listener.Addr().String(),
		message,
		time.Second,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("fake server failed: %v", err)
	}
}

func TestSendAndWaitForOKRejectsUnexpectedResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				serverErr <- err
				return
			}

			serverErr <- fmt.Errorf("client closed connection without sending a message")
			return
		}

		response := protocol.Message{
			Type: protocol.MessageTypeValue,
			From: "server",
		}

		if err := writeMessage(conn, response); err != nil {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	n := New("client", 0)

	message := protocol.Message{
		Type:  protocol.MessageTypeSet,
		From:  n.ID,
		Key:   "name",
		Value: "alice",
	}

	err = n.SendAndWaitForOK(
		listener.Addr().String(),
		message,
		time.Second,
	)

	if err == nil {
		t.Fatal("expected error for unexpected response, got nil")
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("fake server failed: %v", err)
	}
}

func TestPersistentNodeRecoversState(t *testing.T) {
	walPath := filepath.Join(
		t.TempDir(),
		"node-1",
		"wal.log",
	)

	node1, err := NewPersistent(
		"node-1",
		8080,
		walPath,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := node1.applyCommand(
		command.Command{
			Type:  command.TypeSet,
			Key:   "name",
			Value: "alice",
		},
		true,
	); err != nil {
		t.Fatal(err)
	}

	if err := node1.applyCommand(
		command.Command{
			Type:  command.TypeSet,
			Key:   "temporary",
			Value: "value",
		},
		true,
	); err != nil {
		t.Fatal(err)
	}

	if err := node1.applyCommand(
		command.Command{
			Type: command.TypeDelete,
			Key:  "temporary",
		},
		true,
	); err != nil {
		t.Fatal(err)
	}

	if err := node1.Close(); err != nil {
		t.Fatal(err)
	}

	node2, err := NewPersistent(
		"node-1",
		8080,
		walPath,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer node2.Close()

	value, ok := node2.Store.Get("name")

	if !ok {
		t.Fatal("expected recovered key name")
	}

	if value != "alice" {
		t.Fatalf(
			"expected alice, got %s",
			value,
		)
	}

	if _, ok := node2.Store.Get("temporary"); ok {
		t.Fatal(
			"expected deleted key to remain deleted after recovery",
		)
	}
}
func TestPersistentMutationDoesNotApplyWhenWALFails(
	t *testing.T,
) {
	walPath := filepath.Join(
		t.TempDir(),
		"wal.log",
	)

	n, err := NewPersistent(
		"node-1",
		8080,
		walPath,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := n.wal.Close(); err != nil {
		t.Fatal(err)
	}

	err = n.applyCommand(
		command.Command{
			Type:  command.TypeSet,
			Key:   "name",
			Value: "alice",
		},
		true,
	)

	if err == nil {
		t.Fatal("expected WAL append to fail")
	}

	if _, ok := n.Store.Get("name"); ok {
		t.Fatal(
			"expected Store to remain unchanged when WAL append fails",
		)
	}
}
