package internal

import (
	"log"
	"time"
	"fmt"
	"math/rand"
)

const heartBeatTime = time.Millisecond * 1000
const heartBeatTimout = time.Millisecond * 3000
const electionTimer = time.Millisecond * 3000

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
func NewNode(dom string, port string) *Node {
	return &Node{
		Domain: dom,
		Port: port,
		status: follower,
		currentTerm: 0,
		stateMachine: *NewStateMachine(),
		votedFor: "",
		heartBeatChannel: make(chan string),
		newStatusChannel: make(chan Status),
	}
}

// worker used to init the network
func tryToAddPeerWorker(paths <-chan string, clients chan<- Client) {
	p := <-paths
	log.Println("Try to connect to", p)
	maxTry := 5
	i := 0
	for i < maxTry {
		c, err := NewClient(p)
		if (err != nil) {
			// log.Println(err)
			time.Sleep(time.Millisecond * 2000)
		} else {
			clients <- *c
			log.Println("Connection established with", p)
			break
		}
		i++
	}
	if (i == maxTry) {
		log.Println("Connection with", p, "failed")
	}
}

func (n *Node) AddPeer(peers ...string) error {
	nbOfPeers := len(peers)
	paths := make(chan string, nbOfPeers)
	clients := make(chan Client, nbOfPeers)
	// launch the workers
	for w := 0; w < nbOfPeers; w++ {
		go tryToAddPeerWorker(paths, clients)
	}
	// add paths to channel
	for _, path := range peers {
		paths <- path
	}
	close(paths);
	for client := range clients {
		n.network = append(n.network, client)
	}
	close(clients)
	fmt.Println("network", n.network)
	return nil
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

func (n *Node) listenHeartBeat()  {
	log.Println("Start listening for heartbeat")
	for {
		timer := time.NewTimer(heartBeatTimout)
		go func() {
			<-timer.C
			if n.status == follower {
				log.Println("Heartbeat time out. the node becomes candidate")
				n.changeStatus(candidate)
			}
		}()
		receivedFrom := <- n.heartBeatChannel
		log.Println("Heartbeat received by", receivedFrom)
		timer.Stop() 
		if (n.status == candidate) {
			log.Println("A leader has been elected, the node becomes a follower")
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

		timer := time.NewTimer(electionTimer)
		
		// launch workers			
		for range n.network {		
			go n.requestVoteWorker(chanNetwork, chanVotes)
		}

		for _, c := range n.network {
			chanNetwork <- c
		}
		close(chanNetwork)

		go func() {
			<-timer.C
			close(chanVotes)
			go n.startElection()
		}()

		

		for granted := range chanVotes {
			if granted == true {
				nbOfVotes++
				if nbOfVotes > chanSize / 2 {
					n.changeStatus(leader)
					break
				}
			}
		}
		timer.Stop()
	}
}

func (n *Node) requestVoteWorker(cli <-chan Client, granted chan<- bool) {
	for client := range cli {
		_, ok := client.RequestVote(RequestVoteArgs{
			Term: n.currentTerm,
			CandidateId: n.Path(),
			LastLogIndex: len(n.stateMachine.log) - 1,
			LastLogTerm: n.stateMachine.getLastLogTerm(),
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
				fmt.Println(client)
				go client.AppendEntries(AppendEntriesArgs{
					Term: n.currentTerm,
					LeaderId: n.Path(),
					PrevLogIndex: len(n.stateMachine.log) - 1,
					PrevLogTerm: n.stateMachine.getLastLogTerm(),
					Entries: emptyEntries,
					LeaderCommit: n.stateMachine.commitIndex,
				})
			}
			time.Sleep(heartBeatTime)
	}
	return nil
}

