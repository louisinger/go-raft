package main

import (
	"os"
	"log"
	"github.com/louisinger/go-raft/internal"
)

func main ()  {
	args := os.Args[1:]
	port := args[0]
	peers := args[1:]

	network := make(internal.Network, len(peers))
	for index, path := range peers {
		peerClient, err := internal.NewClient(path)
		if (err != nil) {
			log.Println("Can't connect to", path, "| error:", err)
		} else {
			network[index] = *peerClient 
			log.Println("Connection established with", path)
		}
	}
	
	node := internal.NewNode("localhost", port, network)
	server := node.RPCserver()
	internal.InitAndServe(server)
}