package main

import (
	"fmt"
	"os"

	"github.com/jashrashne/concord/internal/config"
	"github.com/jashrashne/concord/internal/node"

	"github.com/jashrashne/concord/internal/protocol"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("configuration error", err)
		os.Exit(1)
	}

	n := node.New(cfg.NodeID, cfg.Port)

	for _, p := range cfg.Peers {
		n.AddPeer(p)
	}

	if cfg.Ping || cfg.Message != "" {
		targetPeer, ok := n.Peer(cfg.Target)
		if !ok {
			fmt.Printf("unknown target peer: %s\n", cfg.Target)
			os.Exit(1)
		}

		if cfg.Ping {
			if err := n.Ping(targetPeer.Address, cfg.PingTimeout); err != nil {
				fmt.Println("ping error:", err)
				os.Exit(1)
			}
		}

		if cfg.Message != "" {
			message := protocol.Message{
				Type: protocol.MessageTypeMessage,
				From: cfg.NodeID,
				Data: cfg.Message,
			}

			if err := n.Send(targetPeer.Address, message); err != nil {
				fmt.Println("send error:", err)
				os.Exit(1)
			}
		}
	}

	if err := n.Start(); err != nil {
		fmt.Println("node error : ", err)
		os.Exit(1)
	}
}
