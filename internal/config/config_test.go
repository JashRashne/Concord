package config

import (
	"testing"

	"github.com/jashrashne/concord/internal/peer"
)

func TestValidatePeersRejectsDuplicateIDs(t *testing.T) {
	peers := []peer.Peer{
		{
			ID:      "node-2",
			Address: "localhost:8081",
		},
		{
			ID:      "node-2",
			Address: "localhost:8082",
		},
	}

	err := validatePeers("node-1", peers)

	if err == nil {
		t.Fatal("expected duplicate peer ID to return an error")
	}
}
