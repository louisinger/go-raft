package main

import (
	"os"
	"net/rpc"
	"net/http"
	"log"
	"net"
	"encoding/gob"
	"github.com/louisinger/go-raft/internal"
)

func main ()  {
	gob.Register(internal.Put{})
	gob.Register(internal.Delete{})

	args := os.Args[1:]
	port := args[0]
	peers := args[1:]

	node := internal.NewNode("localhost", port)
	server := node.RPCserver()
	addr := "localhost:" + port
	// set up the RPC server
	rpc.Register(server)
	rpc.HandleHTTP()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listen error:", err)
	}

	go server.Node.AddPeer(peers...)
	go server.Node.StartRaft()
	// log the start event
	log.Println("Node is running at", server.Node.Path())
	err = http.Serve(ln, nil)
	if err != nil {
		log.Fatal(err)
	}
}