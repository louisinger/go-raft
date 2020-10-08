package main

import (
	"os"
	"github.com/louisinger/go-raft/internal"
)

func main ()  {
	args := os.Args[1:]
	port := args[0]
	peers := args[1:]

	node := internal.NewNode("localhost", port)
	server := node.RPCserver()
	internal.InitAndServe(server, peers...)
}