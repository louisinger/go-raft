package internal

import ( 
	"fmt"
	"log"
	"encoding/json"
)

type Server struct {
	Node Node
}


// RPC method info
func (server *Server) Info(payload string, reply *string) error {
	switch payload {
	case "network":
		*reply = "network"
	case "domain":
		*reply = server.Node.Domain
	case "state":
		j, _ := json.Marshal(server.Node.stateMachine.state)
		*reply = string(j)
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
		reply.Success = true
		return nil
	}
	
	containLogAtPrevIndex := len(server.Node.stateMachine.log) - 1 > payload.PrevLogIndex
	if containLogAtPrevIndex && payload.PrevLogTerm != server.Node.stateMachine.log[payload.PrevLogIndex].getTerm() {
		reply.Success = false
		return nil
	}
	// remove the potential conflicted entries
	server.Node.stateMachine.removeIfConflicts(payload.Entries, payload.PrevLogIndex + 1)

	// append entries in the log
	server.Node.stateMachine.append(payload.PrevLogIndex, payload.Entries...)

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
	log.Println("request vote received from", payload.CandidateId)
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

type PutEntriesArgs struct {
	Key string
	JsonStr string
}

// PutEntries - use for clients to send new entries to the leader
func (server *Server) PutEntries(payload []PutEntriesArgs, reply *bool) error {
	log.Println("Put commands received.")
	switch server.Node.status {
		case candidate, follower:
			log.Println("Put commands redirected to leader.")
			// redirect to leader
			for _, client := range server.Node.network {
				if client.Path == server.Node.leaderId {
					success, e := client.Put(payload)
					*reply = success 
					return e
				}
			}
		case leader:
			termLeader := server.Node.currentTerm
			var cmds []Command
			for _, put := range payload {
				// deserialize and create Command
				var valueJson Json
				if err := json.Unmarshal([]byte(put.JsonStr), &valueJson); err != nil {
					*reply = false
					return err
				}
				cmd := Put{
					Entry: Entry{Term: termLeader},
					key: put.Key,
					value: valueJson,
				}
				cmds = append(cmds, cmd)
			}

			server.Node.stateMachine.log = append(server.Node.stateMachine.log, cmds...) 
			server.Node.stateMachine.commitIndex += len(cmds)
			if err := server.Node.stateMachine.applyUntilCommitIndex(); err != nil {
				log.Println(err)
				*reply = false
				return err
			}
			*reply = true
			return nil
	}
	*reply = false
	return nil
}