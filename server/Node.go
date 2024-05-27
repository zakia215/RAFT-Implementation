package main

import (
	"fmt"
	"net/rpc"
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
	ELECTION_TIMEOUT_MIN = 2 * time.Second
	ELECTION_TIMEOUT_MAX = 3 * time.Second
	RPC_TIMEOUT          = 500 * time.Millisecond
)

type Node struct {
	Address       Address
	Type          NodeType
	ElectionTerm  int
	AddressList   []Address
	LeaderAddress Address
	Application   map[string]string
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
			go n.SendHeartbeatTo(address)
		}
	}
}

func (n *Node) SendHeartbeatTo(address Address) {
	client, err := rpc.Dial("tcp", address.IPAddress+":"+address.Port)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer client.Close()

	var reply string
	err = client.Call("RPCService.Ping", struct{}{}, &reply)
	if err != nil {
		fmt.Println("Error calling Ping to", address.IPAddress+":"+address.Port, ":", err)
		return
	}
	fmt.Println("Heartbeat sent to", address.IPAddress+":"+address.Port, ":", reply)
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
