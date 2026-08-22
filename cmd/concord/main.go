package main

import (
	"fmt"
	"os"

	"github.com/jashrashne/concord/internal/config"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Println("configuration error", err)
		os.Exit(1)
	}

	fmt.Println("Concord starting...")
	fmt.Println("Node ID:", cfg.NodeID)
	fmt.Println("Port:", cfg.Port)
}
