package internal

import ( 
	"net/http" 
	"net/rpc"
	"io"
	"log"
	"fmt"
)

type Server struct {
	Node Node
}

func InitAndServe(server *Server) error {
	// set isAlive GET endpoint
	http.HandleFunc("/", func (res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})

	// set up the RPC server
	rpc.Register(server)
	rpc.HandleHTTP()
	// log the start event
	log.Println("Node is running at", server.Node.Path())
	// listenAndServe function of the http server 
	log.Fatal(http.ListenAndServe(":" + server.Node.Port, nil))
	return nil
}

// RPC method info
func (server *Server) Info(payload string, reply *string) error {
	switch payload {
	case "network":
		*reply = "network"
	case "domain":
		*reply = server.Node.Domain
	default:
		return fmt.Errorf(payload + " is not a submethod of info.")
	}

	return nil
}

// RPC method NewPeer
func (server *Server) NewPeer(path string, reply *bool) error {
	err := server.Node.AddPeer(path)

	if (err != nil) {
		*reply = false
		return err
	} 

	*reply = true
	return nil
}