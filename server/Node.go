package main

import (
	"common"
	"fmt"
	"math/rand"
	"net/rpc"
	"sync"
	"time"
)

// NodeType definitions
type NodeType string

const (
	LEADER    NodeType = "LEADER"
	FOLLOWER  NodeType = "FOLLOWER"
	CANDIDATE NodeType = "CANDIDATE"
)

// constants for timeout and interval
const (
	HEARTBEAT_INTERVAL   = 1 * time.Second
	ELECTION_TIMEOUT_MIN = 5 * time.Second
	ELECTION_TIMEOUT_MAX = 10 * time.Second
	RPC_TIMEOUT          = 500 * time.Millisecond
)

type Node struct {
	Address       Address
	Type          NodeType
	ElectionTerm  int
	VotedFor      *Address
	AddressList   []Address
	LeaderAddress Address
	Application   map[string]string
	mutex         sync.Mutex
	heartbeatCh   chan bool
	Log           []common.LogEntry
	CommitIndex   int
	LastApplied   int
}

func (n *Node) InitializeLeader() {
	n.Type = LEADER
	n.ElectionTerm++
	n.LeaderAddress = n.Address
	go n.StartHeartbeat()
}

func (n *Node) StartHeartbeat() {
	ticker := time.NewTicker(HEARTBEAT_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		if n.Type != LEADER {
			return
		}
		n.SendHeartbeat()
	}
}

func (n *Node) SendHeartbeat() {
	for _, address := range n.AddressList {
		if address != n.Address {
			go func(address Address) {
				err := n.SendHeartbeatTo(address)
				if err != nil {
					fmt.Println("Heartbeat failed to", address.IPAddress+":"+address.Port, ":", err)
					n.mutex.Lock()
					if n.Type == LEADER {
						fmt.Println("Leader heartbeat failed. Starting election.")
						n.Type = FOLLOWER
						go n.StartElection()
					}
					n.mutex.Unlock()
				}
			}(address)
		}
	}
}

func (n *Node) SendHeartbeatTo(address Address) error {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		return err
	}
	defer client.Close()

	var reply string
	err = client.Call("RPCService.Ping", struct{}{}, &reply)
	if err != nil {
		return err
	}
	fmt.Println("Heartbeat sent to", address.IPAddress+":"+address.Port, ":", reply)
	return nil
}

func (n *Node) GetValue(key string) string {
	return n.Application[key]
}

func (n *Node) SetValue(key string, value string) {
	n.Application[key] = value
}

func (n *Node) AddFollower(address Address) {
	n.AddressList = append(n.AddressList, address)
	fmt.Println("New follower added:", address.IPAddress+":"+address.Port)
}

func (n *Node) StartElection() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Type = CANDIDATE
	n.ElectionTerm++
	n.VotedFor = &n.Address

	votes := 1
	voteCh := make(chan bool, len(n.AddressList))

	for _, address := range n.AddressList {
		if address != n.Address {
			go n.RequestVote(address, voteCh)
		}
	}

	electionTimeout := time.Duration(rand.Intn(int(ELECTION_TIMEOUT_MAX.Seconds()-ELECTION_TIMEOUT_MIN.Seconds()))+int(ELECTION_TIMEOUT_MIN.Seconds())) * time.Second
	timeout := time.After(electionTimeout)

	for votes < len(n.AddressList)/2+1 {
		select {
		case voteGranted := <-voteCh:
			if voteGranted {
				votes++
			}
		case <-timeout:
			fmt.Println("Election timeout. Starting new election.")
			go n.StartElection()
			return
		}
	}

	if votes >= len(n.AddressList)/2+1 {
		fmt.Println("Election won. Becoming leader.")
		n.InitializeLeader()
	}
}

func (n *Node) RequestVote(address Address, voteCh chan bool) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		voteCh <- false
		return
	}
	defer client.Close()

	args := RequestVoteArgs{
		Term:        n.ElectionTerm,
		CandidateID: n.Address,
	}
	var reply RequestVoteReply
	err = client.Call("RPCService.RequestVote", args, &reply)
	if err != nil {
		fmt.Println("Error calling RequestVote:", err)
		voteCh <- false
		return
	}

	voteCh <- reply.VoteGranted
}

func (n *Node) ResetElectionTimer() {
	for {
		electionTimeout := time.Duration(rand.Intn(int(ELECTION_TIMEOUT_MAX.Seconds()-ELECTION_TIMEOUT_MIN.Seconds()))+int(ELECTION_TIMEOUT_MIN.Seconds())) * time.Second
		timer := time.NewTimer(electionTimeout)

		for remaining := electionTimeout; remaining > 0; remaining -= time.Second {
			fmt.Printf("Node %s: Election timeout in %d seconds\n", n.Address.IPAddress+":"+n.Address.Port, remaining/time.Second)
			select {
			case <-time.After(time.Second):
				// Continue countdown
			case <-n.heartbeatCh:
				fmt.Println("Heartbeat received, resetting election timer.")
				if !timer.Stop() {
					<-timer.C
				}
				electionTimeout = time.Duration(rand.Intn(int(ELECTION_TIMEOUT_MAX.Seconds()-ELECTION_TIMEOUT_MIN.Seconds()))+int(ELECTION_TIMEOUT_MIN.Seconds())) * time.Second
				timer.Reset(electionTimeout)
				remaining = electionTimeout
			case <-timer.C:
				remaining = 0
			}
		}

		if n.Type == FOLLOWER {
			fmt.Println("Election timeout. Starting election.")
			go n.StartElection()
		}
	}
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

func (n *Node) replicateLog() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for _, address := range n.AddressList {
		if address != n.Address {
			go n.sendAppendEntries(address)
		}
	}
}