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
	ELECTION_TIMEOUT_MAX = 7 * time.Second
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

type RequestVoteArgs struct {
	Term         int
	CandidateID  Address
	LastLogIndex int
	LastLogTerm  int
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

type AddressListReply struct {
	AddressList []Address
}

func (n *Node) InitializeLeader() {
	n.Type = LEADER
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
						n.RemoveNode(address)
						fmt.Println("Leader heartbeat failed to", address.IPAddress+":"+address.Port, ". Removing node.")
						n.NotifyFollowersOfNewAddressList()
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
	n.NotifyFollowersOfNewAddressList()
}

func (n *Node) UpdateAddressList(addressList []Address) {
	n.AddressList = addressList
	fmt.Println("Address list updated:", n.AddressList)
}

func (n *Node) RemoveNode(address Address) {
	for i, addr := range n.AddressList {
		if addr == address {
			n.AddressList = append(n.AddressList[:i], n.AddressList[i+1:]...)
			fmt.Println("Removed node:", address.IPAddress+":"+address.Port)
			break
		}
	}
}

func (n *Node) StartElection() {
	n.mutex.Lock()
	n.Type = CANDIDATE
	n.ElectionTerm++
	n.VotedFor = &n.Address
	n.mutex.Unlock()

	votes := 1
	voteCh := make(chan bool, len(n.AddressList))
	doneCh := make(chan struct{})

	for _, address := range n.AddressList {
		if address != n.Address {
			go n.RequestVote(address, voteCh)
		}
	}

	electionTimeout := time.Duration(rand.Intn(int(ELECTION_TIMEOUT_MAX.Seconds()-ELECTION_TIMEOUT_MIN.Seconds()))+int(ELECTION_TIMEOUT_MIN.Seconds())) * time.Second
	timeout := time.After(electionTimeout)

	go func() {
		for votes < len(n.AddressList)/2+1 {
			select {
			case voteGranted := <-voteCh:
				if voteGranted {
					votes++
				}
			case <-timeout:
				fmt.Println("Election timeout. Starting new election.")
				go n.StartElection()
				close(doneCh)
				return
			}
		}

		if votes >= len(n.AddressList)/2+1 {
			fmt.Println("Election won. Becoming leader with", votes, "vote(s).")
			n.InitializeLeader()
			n.NotifyFollowersOfNewLeader()
		} else {
			fmt.Println("Election lost. Received", votes, "vote(s).")
			n.mutex.Lock()
			n.Type = FOLLOWER
			n.mutex.Unlock()
		}
		close(doneCh)
	}()

	<-doneCh
}

func (n *Node) RequestVote(address Address, voteCh chan bool) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		voteCh <- false
		n.mutex.Lock()
		n.RemoveNode(address)
		n.mutex.Unlock()
		n.NotifyFollowersOfNewAddressList()
		return
	}
	defer client.Close()

	lastLogIndex := len(n.Log) - 1
	lastLogTerm := -1
	if lastLogIndex >= 0 {
		lastLogTerm = n.Log[lastLogIndex].Term
	}

	args := RequestVoteArgs{
		Term:         n.ElectionTerm,
		CandidateID:  n.Address,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}
	var reply RequestVoteReply
	err = client.Call("RPCService.RequestVote", args, &reply)
	if err != nil {
		fmt.Println("Error calling RequestVote:", err)
		voteCh <- false
		n.mutex.Lock()
		n.RemoveNode(address)
		n.mutex.Unlock()
		n.NotifyFollowersOfNewAddressList()
		return
	}

	if reply.Term > n.ElectionTerm {
		n.mutex.Lock()
		n.ElectionTerm = reply.Term
		n.Type = FOLLOWER
		n.VotedFor = nil
		n.mutex.Unlock()
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
			return
		}
	}
}

func (n *Node) sendAppendEntries(address Address) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		n.mutex.Lock()
		n.RemoveNode(address)
		n.mutex.Unlock()
		n.NotifyFollowersOfNewAddressList()
		return
	}
	defer client.Close()

	args := AppendEntriesArgs{
		Term:         n.ElectionTerm,
		LeaderID:     n.Address,
		PrevLogIndex: len(n.Log) - 1,
		PrevLogTerm:  -1,
		Entries:      []common.LogEntry{},
		LeaderCommit: n.CommitIndex,
	}

	if len(n.Log) > 0 {
		args.PrevLogTerm = n.Log[len(n.Log)-1].Term
		args.Entries = append(args.Entries, n.Log[len(n.Log)-1])
	}

	var reply AppendEntriesReply
	err = client.Call("RPCService.AppendEntries", args, &reply)
	if err != nil {
		fmt.Println("Error calling AppendEntries to", address.IPAddress+":"+address.Port, ":", err)
		return
	}

	if !reply.Success && reply.Term > n.ElectionTerm {
		n.mutex.Lock()
		n.ElectionTerm = reply.Term
		n.Type = FOLLOWER
		fmt.Printf("Node %s: New term %d established, now following node %s\n", n.Address, n.ElectionTerm, args.LeaderID)
		n.mutex.Unlock()
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

func (n *Node) NotifyFollowersOfNewLeader() {
	for _, address := range n.AddressList {
		if address != n.Address {
			go n.SendNewLeaderNotification(address)
		}
	}
}

func (n *Node) SendNewLeaderNotification(address Address) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		n.mutex.Lock()
		n.RemoveNode(address)
		n.mutex.Unlock()
		n.NotifyFollowersOfNewAddressList()
		return
	}
	defer client.Close()

	args := AppendEntriesArgs{
		Term:         n.ElectionTerm,
		LeaderID:     n.Address,
		PrevLogIndex: -1, // No log entries
		PrevLogTerm:  -1,
		Entries:      nil,
		LeaderCommit: n.CommitIndex,
	}

	var reply AppendEntriesReply
	err = client.Call("RPCService.AppendEntries", args, &reply)
	if err != nil {
		fmt.Println("Error calling AppendEntries to", address.IPAddress+":"+address.Port, ":", err)
		return
	} else {
		fmt.Println("New leader notification sent to", address.IPAddress+":"+address.Port)
	}

	if !reply.Success && reply.Term > n.ElectionTerm {
		n.mutex.Lock()
		n.ElectionTerm = reply.Term
		n.Type = FOLLOWER
		fmt.Printf("Node %s: New term %d established, now following node %s\n", n.Address, n.ElectionTerm, args.LeaderID)
		n.mutex.Unlock()
	}
}

func (n *Node) NotifyFollowersOfNewAddressList() {
	for _, address := range n.AddressList {
		if address != n.Address {
			go n.SendUpdatedAddressList(address)
		}
	}
}

func (n *Node) SendUpdatedAddressList(address Address) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer client.Close()

	args := n.AddressList
	var reply string
	err = client.Call("RPCService.UpdateAddressList", args, &reply)
	if err != nil {
		fmt.Println("Error calling UpdateAddressList to", address.IPAddress+":"+address.Port, ":", err)
		return
	}
	fmt.Println("Updated address list sent to", address.IPAddress+":"+address.Port)
}
