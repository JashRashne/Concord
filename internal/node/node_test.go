package node

import (
	"testing"

	"github.com/jashrashne/concord/internal/peer"
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
