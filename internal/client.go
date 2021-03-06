package internal

import ( 
	"net/rpc" 
	"log"
)

type Client struct {
	ClientRPC *rpc.Client
	Path string
}

func NewClient(path string) (Client, error) {
	client, err :=	rpc.DialHTTP("tcp", path)
	c := Client{
		Path: path,
		ClientRPC: client,
	}
	if err != nil {
		log.Println("dialing:", err)
		return c, err
	}
	return c, nil
}

// RPC using to retreive information about the node.
func (c *Client) Info(payload string) string {
	var result string
	if err := c.ClientRPC.Call("Server.Info", payload, &result); err != nil {
		log.Println(err)
	}
	return result
}

// RPC using to add a new node into the node's network.
func (c *Client) NewPeer(payload string) bool {
	var result bool
	if err := c.ClientRPC.Call("Server.NewPeer", payload, &result); err != nil {
		log.Println(err)
	}
	return result
}

// Append Entries RPC
func (c *Client) AppendEntries(args AppendEntriesArgs) (int, bool) {
	var result AppendEntriesReply
	if err := c.ClientRPC.Call("Server.AppendEntries", args, &result); err != nil {
		log.Println(err)
	}
	return result.Term, result.Success
}

// Request Vote RPC
func (c *Client) RequestVote(args RequestVoteArgs) (int, bool) {
	var result RequestVoteReply
	if err := c.ClientRPC.Call("Server.RequestVote", args, &result); err != nil {
		log.Println(err)
	}
	return result.Term, result.VoteGranted
}

// PutEntries RPC
func (c *Client) Put(payload []PutEntriesArgs) (bool, error) {
	var result bool
	if err := c.ClientRPC.Call("Server.PutEntries", payload, &result); err != nil {
		log.Println(err)
		return result, err
	}
	return result, nil
}

