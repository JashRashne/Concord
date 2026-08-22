package config

import (
	"errors"
	"flag"
)

type Config struct {
	NodeID string
	Port   int
}

func Parse() (Config, error) {
	nodeID := flag.String("node-id", "node-1", "unique ID for this Concord node")
	port := flag.Int("port", 8080, "port for Concord to listen on")

	flag.Parse()

	cfg := Config{
		NodeID: *nodeID,
		Port:   *port,
	}

	if cfg.NodeID == "" {
		return Config{}, errors.New("node ID cannot be empty")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return Config{}, errors.New("port must be between 1 and 65535")
	}

	return cfg, nil
}
