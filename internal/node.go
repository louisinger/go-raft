package internal

import (
	"log"
	"time"
	"fmt"
	"math/rand"
)

const heartBeatTime = time.Millisecond * 1000
const electionTimer = time.Millisecond * 5000

// status of node (enum)
type Status string
const (
	follower Status = "follower"
	candidate Status = "candidate"
	leader Status = "leader"
)

type Node struct {
	Domain string
	Port string
	status Status
	network Network
	currentTerm int
	stateMachine StateMachine
	votedFor string
	heartBeatChannel chan string
	newStatusChannel chan Status
}

type Network []Client

// factory function for Node structure.
func NewNode(dom string, port string, net Network) *Node {
	return &Node{
		Domain: dom,
		Port: port,
		status: follower,
		network: net,
		currentTerm: 0,
		stateMachine: *NewStateMachine(),
		votedFor: "",
		heartBeatChannel: make(chan string),
		newStatusChannel: make(chan Status),
	}
}

// create an RPC Server from a node.
func (n Node) RPCserver() *Server {
	return &Server{
		Node: n,
	}
}

func (n *Node) changeStatus(st Status) {
	n.status = st
	n.newStatusChannel <- st
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

func (n *Node) listenHeartBeat()  {
	log.Println("Start listening for heartbeat")
	for {
		timer := time.NewTimer(heartBeatTime)
		go func() {
			<-timer.C
			if n.status == follower {
				log.Println("Heartbeat time out. Node become candidate.")
				n.changeStatus(candidate)
			}
		}()
		<- n.heartBeatChannel
		timer.Stop() 
		if (n.status == candidate) {
			n.changeStatus(follower)
		}
	}
}

func (n *Node) StartRaft() error {
	log.Println("Start raft algorithm...")
	go n.listenHeartBeat()
	for { 
		log.Println("Launch work depending of status:", n.status)
		switch (n.status) {
			case leader:
				go n.sendHeartBeat()
			case candidate:
				// timerElection := time.NewTimer(electionTimer)
				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
				go n.startElection()
		}
		<-n.newStatusChannel 
	}
}

func (n *Node) startElection() {
		log.Println("Node begins election.")
		nbOfVotes := 0
		n.votedFor = n.Path()

		chanSize := len(n.network)

		if chanSize == 0 {
			n.changeStatus(leader)
		} else {
			chanNetwork := make(chan Client, chanSize)
			chanVotes := make(chan bool, chanSize)

			go n.requestVoteWorker(chanNetwork, chanVotes)
			go n.requestVoteWorker(chanNetwork, chanVotes)

			for _, c := range n.network {
				chanNetwork <- c
			}
			close(chanNetwork)
			for granted := range chanVotes {
				if granted == true {
					nbOfVotes++
					if nbOfVotes > chanSize / 2 {
						n.changeStatus(leader)
						break
					}
				}
			}
			close(chanVotes)
			n.changeStatus(follower)
		}
}

func (n *Node) requestVoteWorker(cli <-chan Client, granted chan<- bool) {
	for client := range cli {
		_, ok := client.RequestVote(RequestVoteArgs{
			term: n.currentTerm,
			candidateId: n.Path(),
			lastLogIndex: len(n.stateMachine.log) - 1,
			lastLogTerm: n.stateMachine.getLastLogTerm(),
		})

		granted<-ok
	}
}

// send empty appendEntries RPC to all nodes
func (n *Node) sendHeartBeat() error {
		var emptyEntries = []Command{}

		if n.status != leader {
			return fmt.Errorf("The node is not the leader, so it can't begin to send heartbeat")
		}
		for n.status == leader {
			log.Println("Send heartbeat to", len(n.network), "nodes")
			// send AppendEntries empty to all client
			for _, client := range n.network {
				go client.AppendEntries(AppendEntriesArgs{
					term: n.currentTerm,
					leaderId: n.Path(),
					prevLogIndex: len(n.stateMachine.log) - 1,
					prevLogTerm: n.stateMachine.getLastLogTerm(),
					entries: emptyEntries,
					leaderCommit: n.stateMachine.commitIndex,
				})
			}
			time.Sleep(heartBeatTime)
	}
	return nil
}

