package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/jashrashne/concord/internal/peer"
)

type peerList []peer.Peer
type Config struct {
	NodeID         string
	Port           int
	Peers          []peer.Peer
	Target         string
	Message        string
	Ping           bool
	RequestTimeout time.Duration

	Set    bool
	Key    string
	Value  string
	Get    bool
	Delete bool
}

func Parse() (Config, error) {
	var peers peerList

	nodeID := flag.String("node-id", "node-1", "unique ID for this Concord node")
	port := flag.Int("port", 8080, "port for Concord to listen on")
	flag.Var(
		&peers,
		"peer",
		"peer in ID=ADDRESS format (repeatable)",
	)
	message := flag.String("message", "", "message to send the peer node")
	ping := flag.Bool("ping", false, "ping the configured peer")
	requestTimeout := flag.Duration(
		"ping-timeout",
		3*time.Second,
		"maximum time to wait for a PONG",
	)
	target := flag.String(
		"target",
		"",
		"ID of the peer to contact",
	)
	set := flag.Bool(
		"set",
		false,
		"set a key/value pair on the target peer",
	)

	key := flag.String(
		"key",
		"",
		"key for a database operation",
	)

	value := flag.String(
		"value",
		"",
		"value for a SET operation",
	)

	get := flag.Bool(
		"get",
		false,
		"get a value from the target peer",
	)
	deleteKey := flag.Bool(
		"delete",
		false,
		"delete a key from the target peer",
	)

	flag.Parse()

	cfg := Config{
		NodeID:         *nodeID,
		Port:           *port,
		Peers:          []peer.Peer(peers),
		Message:        *message,
		Ping:           *ping,
		RequestTimeout: *requestTimeout,
		Target:         *target,

		Set:    *set,
		Key:    *key,
		Value:  *value,
		Get:    *get,
		Delete: *deleteKey,
	}

	if cfg.NodeID == "" {
		return Config{}, errors.New("node ID cannot be empty")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, errors.New("port must be between 1 and 65535")
	}

	if cfg.Ping && cfg.Target == "" {
		return Config{}, errors.New("--ping requires --target")
	}

	if cfg.Message != "" && cfg.Target == "" {
		return Config{}, errors.New("--message requires --target")
	}

	if cfg.Ping && cfg.Message != "" {
		return Config{}, errors.New("--ping and --message cannot be used together")
	}

	if cfg.RequestTimeout <= 0 {
		return Config{}, errors.New("request timeout must be greater than zero")
	}

	if err := validatePeers(cfg.NodeID, cfg.Peers); err != nil {
		return Config{}, err
	}

	if cfg.Set && cfg.Target == "" {
		return Config{}, errors.New("--set requires --target")
	}

	if cfg.Set && cfg.Key == "" {
		return Config{}, errors.New("--set requires --key")
	}

	if cfg.Set && (cfg.Ping || cfg.Message != "") {
		return Config{}, errors.New("--set cannot be used with --ping or --message")
	}

	if cfg.Get && cfg.Target == "" {
		return Config{}, errors.New("--get requires --target")
	}

	if cfg.Get && cfg.Key == "" {
		return Config{}, errors.New("--get requires --key")
	}

	if cfg.Get && (cfg.Set || cfg.Ping || cfg.Message != "") {
		return Config{}, errors.New("--get cannot be used with --set, --ping, or --message")
	}

	if cfg.Delete && cfg.Target == "" {
		return Config{}, errors.New("--delete requires --target")
	}

	if cfg.Delete && cfg.Key == "" {
		return Config{}, errors.New("--delete requires --key")
	}

	if cfg.Delete && (cfg.Set || cfg.Get || cfg.Ping || cfg.Message != "") {
		return Config{}, errors.New("--delete cannot be used with --set, --get, --ping, or --message")
	}

	return cfg, nil
}

func validatePeers(nodeID string, peers []peer.Peer) error {
	seenIDs := make(map[string]bool)
	seenAddresses := make(map[string]bool)

	for _, p := range peers {
		if p.ID == nodeID {
			return fmt.Errorf("node %q cannot be its own peer", nodeID)
		}

		if seenIDs[p.ID] {
			return fmt.Errorf("duplicate peer ID %q", p.ID)
		}

		if seenAddresses[p.Address] {
			return fmt.Errorf("duplicate peer address %q", p.Address)
		}

		seenIDs[p.ID] = true
		seenAddresses[p.Address] = true
	}

	return nil
}

func (p *peerList) Set(value string) error {
	id, address, found := strings.Cut(value, "=")

	if !found || id == "" || address == "" {
		return fmt.Errorf("peer must be in ID=ADDRESS format")
	}

	*p = append(*p, peer.Peer{
		ID:      id,
		Address: address,
	})

	return nil
}

func (p *peerList) String() string {
	if p == nil {
		return ""
	}

	parts := make([]string, 0, len(*p))

	for _, currentPeer := range *p {
		parts = append(
			parts,
			fmt.Sprintf("%s=%s", currentPeer.ID, currentPeer.Address),
		)
	}

	return strings.Join(parts, ",")
}
