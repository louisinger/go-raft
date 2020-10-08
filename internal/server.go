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

func InitAndServe(server *Server, initPeers ...string) error {
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
	go server.Node.AddPeer(initPeers...)
	go server.Node.StartRaft()
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
	Term int
	LeaderId string
	PrevLogIndex int
	PrevLogTerm int
	Entries []Command
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term int
	Success bool
}

// RPC method AppendEntries
func (server *Server) AppendEntries(payload AppendEntriesArgs, reply *AppendEntriesReply) error {
	reply.Term = server.Node.currentTerm 
	if payload.Term < server.Node.currentTerm {
		reply.Success = false
		return nil
	}

	if len(payload.Entries) == 0 {
		server.Node.heartBeatChannel <- payload.LeaderId
	}
	
	containLogAtPrevIndex := len(server.Node.stateMachine.log) - 1 > payload.PrevLogIndex
	if containLogAtPrevIndex && payload.PrevLogTerm != server.Node.stateMachine.log[payload.PrevLogIndex].getTerm() {
		reply.Success = false
		return nil
	}
	// remove the potential conflicted entries
	server.Node.stateMachine.removeIfConflicts(payload.Entries, payload.PrevLogIndex + 1)

	// append entries in the log
	server.Node.stateMachine.append(payload.PrevLogIndex + 1, payload.Entries...)

	// set the commit index if leader's commit > node's commit
	if (payload.LeaderCommit > server.Node.stateMachine.commitIndex) {
		server.Node.stateMachine.setCommitIndex(payload.LeaderCommit)
	}
	return nil
}

type RequestVoteArgs struct {
	Term int
	CandidateId string
	LastLogIndex int
	LastLogTerm int
}

type RequestVoteReply struct {
	Term int
	VoteGranted bool
}


// RPC method RequestVoteRPC
func (server *Server) RequestVote(payload RequestVoteArgs, reply *RequestVoteReply) error {
	reply.Term = server.Node.currentTerm 

	if payload.Term < server.Node.currentTerm {
		reply.VoteGranted = false
		return nil
	}
	// check the value of votedFor	
	votedForIsOk := server.Node.votedFor == "" || server.Node.votedFor == payload.CandidateId
	// if the the node has granted a vote before, return false
	if votedForIsOk == false {
		reply.VoteGranted = false
		return nil
	}

	candidateIsUpToDate := (server.Node.stateMachine.getLastLogTerm() <= payload.LastLogTerm) && (len(server.Node.stateMachine.log)-1 <= payload.LastLogIndex)

	if candidateIsUpToDate == false {
		reply.VoteGranted = false
		return nil
	}

	reply.VoteGranted = true
	return nil
}
