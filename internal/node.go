package internal

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
	state map[string]string
	log []string
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

func (n Node) Path() string {
	return n.Domain + ":" + n.Port
}

func (n *Node) AddPeer(path string) error {
	peerClient, err := NewClient(path)
	if (err != nil) {
		return err
	} else {
		n.network = append(n.network, *peerClient)
		return nil
	}
}

