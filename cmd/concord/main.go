package main

import (
	"flag"
	"fmt"
)

func main() {
	port := flag.Int("port", 8080, "port for Concord to listen on")

	flag.Parse()

	fmt.Println("Concord starting...")
	fmt.Println("Port:", *port)
}
