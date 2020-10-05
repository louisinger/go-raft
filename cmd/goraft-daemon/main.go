package main

import (
	"os"
	"strings"
	"github.com/louisinger/go-raft/internal"
)

func main ()  {
	args := os.Args[1:]
	port := args[0]
	peers := args[1:]

	network := make(internal.Network, len(peers))
	for index, peer := range peers {
		s:= strings.Split(peer, ":")
		peerNode := internal.NewNode(s[0], s[1], nil)
		network[index] = *peerNode
	}
	
	node := internal.NewNode("localhost", port, network)
	server := node.RPCserver()
	internal.InitAndServe(server)
}