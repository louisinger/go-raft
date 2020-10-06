package internal

import ( 
	"net/rpc" 
	"log"
)

type Client struct {
	ClientRPC *rpc.Client
}

func NewClient(path string) (*Client, error) {
	client, err :=	rpc.DialHTTP("tcp", path)
	if err != nil {
		return nil, err
	}
	c := &Client{
		ClientRPC: client,
	}

	return c, nil
}

func (c *Client) Info(payload string) string {
	var result string
	if err := c.ClientRPC.Call("Server.Info", payload, &result); err != nil {
		log.Println(err)
	}
	return result
}

func (c *Client) NewPeer(payload string) bool {
	var result bool
	if err := c.ClientRPC.Call("Server.NewPeer", payload, &result); err != nil {
		log.Println(err)
	}
	return result
}
