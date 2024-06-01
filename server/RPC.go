package main

import (
	"common"
	"errors"
	"fmt"
	"strconv"
)

type RPCService struct {
	Node *Node
}

func (r *RPCService) Ping(_ struct{}, reply *string) error {
	fmt.Println("PING")
	*reply = "PONG"
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

func (r *RPCService) GetLog(_ struct{}, reply *[]common.LogEntry) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	*reply = r.Node.Log
	return nil
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

func (r *RPCService) AddFollower(address Address, reply *AddressListReply) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	command := fmt.Sprintf("ADDFOLLOWER %s:%s", address.IPAddress, address.Port)
	logEntry := common.LogEntry{
		Term:    r.Node.ElectionTerm,
		Command: command,
	}
	r.Node.Log = append(r.Node.Log, logEntry)

	r.Node.AddFollower(address)
	reply.AddressList = r.Node.AddressList
	r.Node.NotifyFollowersOfNewAddressList()
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

	if args.Term > r.Node.ElectionTerm {
		r.Node.ElectionTerm = args.Term
		r.Node.VotedFor = nil
	}

	lastLogIndex := len(r.Node.Log) - 1
	lastLogTerm := -1
	if lastLogIndex >= 0 {
		lastLogTerm = r.Node.Log[lastLogIndex].Term
	}

	if (r.Node.VotedFor == nil || *r.Node.VotedFor == args.CandidateID) &&
		(args.LastLogTerm > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)) {
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
	r.Node.Type = FOLLOWER
	r.Node.heartbeatCh <- true // Reset the election timer

	if len(r.Node.Log) > args.PrevLogIndex && args.PrevLogIndex >= 0 && r.Node.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
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
	fmt.Printf("Node %s: New term %d established, now following node %s\n", r.Node.Address, r.Node.ElectionTerm, args.LeaderID)
	return nil
}

func (r *RPCService) UpdateAddressList(addressList []Address, reply *string) error {
	r.Node.mutex.Lock()
	defer r.Node.mutex.Unlock()
	r.Node.UpdateAddressList(addressList)
	*reply = "Address list updated"
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
