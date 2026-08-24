package main

import (
	"fmt"
	"os"

	"github.com/jashrashne/concord/internal/config"
	"github.com/jashrashne/concord/internal/node"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("configuration error", err)
		os.Exit(1)
	}

	n := node.New(cfg.NodeID, cfg.Port)

	if cfg.Peer != "" && cfg.Message != "" {
		if err := n.Send(cfg.Peer, cfg.Message); err != nil {
			fmt.Println("error while sending: ", err)
			os.Exit(1)
		}
	}

	if err := n.Start(); err != nil {
		fmt.Println("node error : ", err)
		os.Exit(1)
	}
}
