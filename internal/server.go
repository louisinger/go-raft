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

type AppendEntriesArgs struct {
	term int
	leaderId string
	prevLogIndex int
	prevLogTerm int
	entries []Command
	leaderCommit int
}

type AppendEntriesReply struct {
	term int
	success bool
}

// RPC method AppendEntries
func (server *Server) AppendEntries(payload AppendEntriesArgs, reply *AppendEntriesReply) error {
	reply.term = server.Node.currentTerm 
	if (payload.term < server.Node.currentTerm) {
		reply.success = false
		return nil
	}
	
	containLogAtPrevIndex := len(server.Node.stateMachine.log) - 1 > payload.prevLogIndex
	if (containLogAtPrevIndex && payload.prevLogTerm != server.Node.stateMachine.log[payload.prevLogIndex].getTerm()) {
		reply.success = false
		return nil
	}
	// remove the potential conflicted entries
	server.Node.stateMachine.removeIfConflicts(payload.entries, payload.prevLogIndex + 1)

	// append entries in the log
	server.Node.stateMachine.append(payload.prevLogIndex + 1, payload.entries...)

	// set the commit index if leader's commit > node's commit
	if (payload.leaderCommit > server.Node.stateMachine.commitIndex) {
		server.Node.stateMachine.setCommitIndex(payload.leaderCommit)
	}
	return nil
}

type RequestVoteArgs struct {
	term int
	candidateId string
	lastLogIndex int
	lastLogTerm int
}

type RequestVoteReply struct {
	term int
	voteGranted bool
}


// RPC method RequestVoteRPC
func (server *Server) RequestVote(payload RequestVoteArgs, reply *RequestVoteReply) error {
	reply.term = server.Node.currentTerm 

	if payload.term < server.Node.currentTerm {
		reply.voteGranted = false
		return nil
	}
	// check the value of votedFor	
	votedForIsOk := server.Node.votedFor == "" || server.Node.votedFor == payload.candidateId
	// if the the node has granted a vote before, return false
	if votedForIsOk == false {
		reply.voteGranted = false
		return nil
	}

	candidateIsUpToDate := (server.Node.stateMachine.getLastLogTerm() <= payload.lastLogTerm) && (len(server.Node.stateMachine.log)-1 <= payload.lastLogIndex)

	if candidateIsUpToDate == false {
		reply.voteGranted = false
		return nil
	}

	reply.voteGranted = true
	return nil
}
