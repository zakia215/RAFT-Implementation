package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type RequestVoteArgs struct {
	Term         int
	CandidateId  string
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type ExecuteArgs struct {
	Command string
	Key     string
	Value   string
}

type ExecuteReply struct {
	Response string
}

type JoinArgs struct {
	Id string
}

type JoinReply struct {
	Peers []string
}

func (n *Node) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.CurrentTerm {
		log.Printf("Received higher term %d, updating current term from %d to %d and converting to follower", args.Term, n.CurrentTerm, args.Term)
		n.CurrentTerm = args.Term
		n.VotedFor = nil
		n.State = Follower
	}

	reply.Term = n.CurrentTerm

	if args.Term < n.CurrentTerm || (n.VotedFor != nil && *n.VotedFor != args.CandidateId) {
		reply.VoteGranted = false
		return nil
	}

	lastLogIndex := len(n.Log) - 1
	lastLogTerm := 0
	if len(n.Log) > 0 {
		lastLogTerm = n.Log[lastLogIndex].Term
	}

	if args.LastLogTerm > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex) {
		n.VotedFor = &args.CandidateId
		reply.VoteGranted = true
		log.Printf("Granting vote to candidate %s for term %d", args.CandidateId, args.Term)
		n.resetElectionTimer()
	} else {
		reply.VoteGranted = false
	}
	return nil
}

func (n *Node) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.CurrentTerm {
		log.Printf("Received higher term %d, updating current term from %d to %d and converting to follower", args.Term, n.CurrentTerm, args.Term)
		n.CurrentTerm = args.Term
		n.VotedFor = nil
		n.State = Follower
		n.resetElectionTimer()
	}

	reply.Term = n.CurrentTerm

	if args.Term < n.CurrentTerm {
		reply.Success = false
		return nil
	}

	// Check if log contains an entry at PrevLogIndex with PrevLogTerm
	if args.PrevLogIndex >= 0 && (args.PrevLogIndex >= len(n.Log) || n.Log[args.PrevLogIndex].Term != args.PrevLogTerm) {
		reply.Success = false
		return nil
	}

	// Append new entries
	for i, entry := range args.Entries {
		if len(n.Log) <= args.PrevLogIndex+1+i {
			n.Log = append(n.Log, entry)
		} else if n.Log[args.PrevLogIndex+1+i].Term != entry.Term {
			n.Log = n.Log[:args.PrevLogIndex+1+i]
			n.Log = append(n.Log, entry)
		}
	}

	if args.LeaderCommit > n.CommitIndex {
		n.CommitIndex = min(args.LeaderCommit, len(n.Log)-1)
	}

	reply.Success = true
	log.Println("AppendEntries successful, resetting election timeout")
	n.resetElectionTimer()
	return nil
}

func (n *Node) Execute(args ExecuteArgs, reply *ExecuteReply) error {
	log.Printf("Received command %s", args.Command)
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.State != Leader {
		reply.Response = "NOT LEADER"
		return nil
	}

	switch args.Command {
	case "ping":
		reply.Response = "PONG"
	case "get":
		if value, ok := n.Store[args.Key]; ok {
			reply.Response = value
		} else {
			reply.Response = ""
		}
	case "set":
		n.Store[args.Key] = args.Value
		reply.Response = "OK"
	case "strln":
		if value, ok := n.Store[args.Key]; ok {
			reply.Response = fmt.Sprintf("%d", len(value))
		} else {
			reply.Response = "0"
		}
	case "del":
		if value, ok := n.Store[args.Key]; ok {
			delete(n.Store, args.Key)
			reply.Response = value
		} else {
			reply.Response = ""
		}
	case "append":
		n.Store[args.Key] += args.Value
		reply.Response = "OK"
	default:
		reply.Response = "UNKNOWN COMMAND"
	}
	return nil
}

func (n *Node) GetLog(args struct{}, reply *[]LogEntry) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Println("GetLog called")

	*reply = n.Log
	return nil
}

func (n *Node) Join(args JoinArgs, reply *JoinReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Add new peer to the list
	for _, peer := range n.Peers {
		if peer == args.Id {
			// Peer already exists
			reply.Peers = n.Peers
			return nil
		}
	}

	n.Peers = append(n.Peers, args.Id)
	log.Printf("Peer %s joined the network", args.Id)

	// Notify all existing peers about the new peer
	for _, peer := range n.Peers {
		if peer == args.Id || peer == n.Id {
			continue
		}

		go func(peer string) {
			client, err := rpc.DialHTTP("tcp", peer)
			if err != nil {
				log.Printf("Failed to connect to peer %s: %v", peer, err)
				return
			}
			defer client.Close()

			joinArgs := JoinArgs{Id: args.Id}
			var joinReply JoinReply

			log.Printf("Notifying peer %s about new peer %s", peer, args.Id)
			err = client.Call("Node.Join", joinArgs, &joinReply)
			if err != nil {
				log.Printf("Join RPC call to peer %s failed: %v", peer, err)
			}
		}(peer)
	}

	// Return the updated list of peers to the new node
	reply.Peers = n.Peers
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func call(peer string, rpcName string, args interface{}, reply interface{}) bool {
	client, err := rpc.DialHTTP("tcp", peer)
	if err != nil {
		log.Printf("Failed to connect to peer %s: %v", peer, err)
		return false
	}
	defer client.Close()

	err = client.Call(rpcName, args, reply)
	if err != nil {
		log.Printf("RPC call %s to peer %s failed: %v", rpcName, peer, err)
		return false
	}
	return true
}
