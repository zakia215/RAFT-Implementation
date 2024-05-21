package main

// NodeType definitions
type NodeType string

const (
	LEADER    NodeType = "LEADER"
	FOLLOWER  NodeType = "FOLLOWER"
	CANDIDATE NodeType = "CANDIDATE"
)

// constants for timeout and interval
const (
	HEARTBEAT_INTERVAL   float32 = 1
	ELECTION_TIMEOUT_MIN float32 = 2
	ELECTION_TIMEOUT_MAX float32 = 3
	RPC_TIMEOUT          float32 = 0.5
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
	// TODO: make a node as leader
}

func (n *Node) SendHeartbeat() {
	// TODO: send a heartbeat to followers
}

func (n *Node) GetValue(key string) string {
	return n.Application[key]
}

func (n *Node) SetValue(key string, value string) {
	n.Application[key] = value
}
