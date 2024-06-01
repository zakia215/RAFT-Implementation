package main

import (
	"common"
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
)

type RPCService struct {
	Node *Node
}

type RequestVoteArgs struct {
	Term        int
	CandidateID Address
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     Address
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []common.LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (r *RPCService) Ping(_ struct{}, reply *string) error {
	fmt.Println("CYKA")
	*reply = "BLYAT"
	r.Node.heartbeatCh <- true
	return nil
}

func (r *RPCService) Get(key string, reply *string) error {
	fmt.Println("GET ", key)
	*reply = r.Node.GetValue(key)
	return nil
}

func (r *RPCService) Set(args []string, reply *string) error {
	if len(args) < 2 {
		return errors.New("insufficient number of arguments")
	}
	key := args[0]
	val := args[1]
	command := fmt.Sprintf("SET %s %s", key, val)

	r.Node.mutex.Lock()
	logEntry := common.LogEntry{
		Term:    r.Node.ElectionTerm,
		Command: command,
	}
	r.Node.Log = append(r.Node.Log, logEntry)
	r.Node.mutex.Unlock()

	if r.Node.Type == LEADER {
		r.Node.replicateLog()
	}

	r.Node.Application[key] = val
	*reply = "OK"
	return nil
}

func (n *Node) replicateLog() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for _, address := range n.AddressList {
		if address != n.Address {
			go n.sendAppendEntries(address)
		}
	}
}

func (r *RPCService) GetLog(_ struct{}, reply *[]common.LogEntry) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	*reply = r.Node.Log
	return nil
}

func (n *Node) sendAppendEntries(address Address) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer client.Close()

	args := AppendEntriesArgs{
		Term:         n.ElectionTerm,
		LeaderID:     n.Address,
		PrevLogIndex: len(n.Log) - 1,
		PrevLogTerm:  n.Log[len(n.Log)-1].Term,
		Entries:      []common.LogEntry{n.Log[len(n.Log)-1]},
		LeaderCommit: n.CommitIndex,
	}

	var reply AppendEntriesReply
	err = client.Call("RPCService.AppendEntries", args, &reply)
	if err != nil {
		fmt.Println("Error calling AppendEntries to", address.IPAddress+":"+address.Port, ":", err)
		return
	}

	if !reply.Success && reply.Term > n.ElectionTerm {
		n.ElectionTerm = reply.Term
		n.Type = FOLLOWER
	}
}

func (r *RPCService) Strln(key string, reply *string) error {
	fmt.Println("STRLN " + key)
	*reply = strconv.Itoa(len(r.Node.GetValue(key)))
	return nil
}

func (r *RPCService) Del(key string, reply *string) error {
	fmt.Println("DEL " + key)

	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	command := fmt.Sprintf("DEL %s", key)
	logEntry := common.LogEntry{
		Term:    r.Node.ElectionTerm,
		Command: command,
	}
	r.Node.Log = append(r.Node.Log, logEntry)

	*reply = r.Node.GetValue(key)
	delete(r.Node.Application, key)
	return nil
}

func (r *RPCService) Append(args []string, reply *string) error {
	if len(args) < 2 {
		return errors.New("insufficient number of arguments")
	}

	key := args[0]
	value := args[1]
	command := fmt.Sprintf("APPEND %s %s", key, value)
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	logEntry := common.LogEntry{
		Term:    r.Node.ElectionTerm,
		Command: command,
	}
	r.Node.Log = append(r.Node.Log, logEntry)

	r.Node.SetValue(args[0], r.Node.GetValue(args[0])+args[1])
	*reply = "OK"
	return nil
}

func (r *RPCService) AddFollower(address Address, reply *string) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	command := fmt.Sprintf("ADDFOLLOWER %s:%s", address.IPAddress, address.Port)
	logEntry := common.LogEntry{
		Term:    r.Node.ElectionTerm,
		Command: command,
	}
	r.Node.Log = append(r.Node.Log, logEntry)

	r.Node.AddFollower(address)
	*reply = "Follower added: " + address.IPAddress + ":" + address.Port
	return nil
}

func (r *RPCService) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()

	if args.Term < r.Node.ElectionTerm {
		reply.Term = r.Node.ElectionTerm
		reply.VoteGranted = false
		return nil
	}

	if r.Node.VotedFor == nil || *r.Node.VotedFor == args.CandidateID {
		r.Node.VotedFor = &args.CandidateID
		reply.VoteGranted = true
	} else {
		reply.VoteGranted = false
	}

	reply.Term = r.Node.ElectionTerm
	return nil
}

func (r *RPCService) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()

	if args.Term < r.Node.ElectionTerm {
		reply.Term = r.Node.ElectionTerm
		reply.Success = false
		return nil
	}

	r.Node.ElectionTerm = args.Term
	r.Node.LeaderAddress = args.LeaderID

	if len(r.Node.Log) > args.PrevLogIndex && r.Node.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Term = r.Node.ElectionTerm
		reply.Success = false
		return nil
	}

	r.Node.Log = append(r.Node.Log[:args.PrevLogIndex+1], args.Entries...)
	if args.LeaderCommit > r.Node.CommitIndex {
		r.Node.CommitIndex = min(args.LeaderCommit, len(r.Node.Log)-1)
	}

	reply.Term = r.Node.ElectionTerm
	reply.Success = true
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
