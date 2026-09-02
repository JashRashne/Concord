package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	commandMode :=
		cfg.Ping ||
			cfg.Message != "" ||
			cfg.Set ||
			cfg.Get ||
			cfg.Delete

	var n *node.Node

	if commandMode {
		n = node.New(cfg.NodeID, cfg.Port)
	} else {
		walPath := filepath.Join(
			cfg.DataDir,
			cfg.NodeID,
			"wal.log",
		)

		n, err = node.NewPersistent(
			cfg.NodeID,
			cfg.Port,
			walPath,
		)
		if err != nil {
			fmt.Println("node recovery error:", err)
			os.Exit(1)
		}

		defer n.Close()
	}

	for _, p := range cfg.Peers {
		n.AddPeer(p)
	}

	if cfg.Ping || cfg.Message != "" || cfg.Set || cfg.Get || cfg.Delete {
		targetPeer, ok := n.Peer(cfg.Target)
		if !ok {
			fmt.Printf("unknown target peer: %s\n", cfg.Target)
			os.Exit(1)
		}

		if cfg.Ping {
			if err := n.Ping(targetPeer.Address, cfg.RequestTimeout); err != nil {
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

		if cfg.Set {
			message := protocol.Message{
				Type:  protocol.MessageTypeSet,
				From:  cfg.NodeID,
				Key:   cfg.Key,
				Value: cfg.Value,
			}

			if err := n.SendAndWaitForOK(
				targetPeer.Address,
				message,
				cfg.RequestTimeout,
			); err != nil {
				fmt.Println("set error:", err)
				os.Exit(1)
			}
		}

		if cfg.Get {
			value, ok, err := n.Get(
				targetPeer.Address,
				cfg.Key,
				cfg.RequestTimeout,
			)
			if err != nil {
				fmt.Println("get error:", err)
				os.Exit(1)
			}

			if !ok {
				fmt.Println("NOT_FOUND")
				return
			}

			fmt.Printf("VALUE %s\n", value)
		}

		if cfg.Delete {
			message := protocol.Message{
				Type: protocol.MessageTypeDelete,
				From: cfg.NodeID,
				Key:  cfg.Key,
			}

			if err := n.SendAndWaitForOK(
				targetPeer.Address,
				message,
				cfg.RequestTimeout,
			); err != nil {
				fmt.Println("delete error:", err)
				os.Exit(1)
			}
		}

		return
	}

	if err := n.Start(); err != nil {
		fmt.Println("node error : ", err)
		os.Exit(1)
	}
}
