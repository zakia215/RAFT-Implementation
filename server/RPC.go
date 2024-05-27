package main

import (
	"errors"
	"fmt"
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

func (r *RPCService) Ping(_ struct{}, reply *string) error {
	fmt.Println("PING")
	*reply = "PONG"
	r.Node.heartbeatCh <- true // Send signal to reset election timer
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
	fmt.Println("SET " + key + " " + val)
	r.Node.SetValue(key, val)
	*reply = "OK"
	return nil
}

func (r *RPCService) Strln(key string, reply *string) error {
	fmt.Println("STRLN " + key)
	*reply = strconv.Itoa(len(r.Node.GetValue(key)))
	return nil
}

func (r *RPCService) Del(key string, reply *string) error {
	fmt.Println("DEL " + key)
	*reply = r.Node.GetValue(key)
	delete(r.Node.Application, key)
	return nil
}

func (r *RPCService) Append(args []string, reply *string) error {
	if len(args) < 2 {
		return errors.New("insufficient number of arguments")
	}
	r.Node.SetValue(args[0], r.Node.GetValue(args[0])+args[1])
	*reply = "OK"
	return nil
}

func (r *RPCService) AddFollower(address Address, reply *string) error {
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