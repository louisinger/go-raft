package internal

import ("log")

// status of node (enum)
type status int
const (
	follower status = 0
	candidate status = 1
	leader status = 2
)

type Node struct {
	Domain string
	Port string
	Status status
	network Network
	stateMachine StateMachine
}

type Network []Client

// factory function for Node structure.
func NewNode(dom string, port string, net Network) *Node {
	return &Node{
		Domain: dom,
		Port: port,
		Status: follower,
		network: net,
	}
}

// create an RPC Server from a node.
func (n Node) RPCserver() *Server {
	return &Server{
		Node: n,
	}
}

// return the path of the node
func (n Node) Path() string {
	return n.Domain + ":" + n.Port
}

func (n *Node) AddPeer(path string) error {
	peerClient, err := NewClient(path)
	if (err != nil) {
		return err
	} else {
		n.network = append(n.network, *peerClient)
		log.Println("Connection established with", path)
		return nil
	}
}

