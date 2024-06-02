package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type LogEntry struct {
	Term    int
	Command interface{}
}

type CommitEntry struct {
	Term    int
	Index   int
	Command interface{}
}

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type Node struct {
	mu                 sync.Mutex
	Id                 string
	State              State
	CurrentTerm        int
	VotedFor           *string
	Log                []LogEntry
	CommitIndex        int
	LastApplied        int
	NextIndex          map[string]int
	MatchIndex         map[string]int
	ElectionTimer      *time.Timer
	HeartbeatTimer     *time.Ticker
	Peers              []string
	VoteCount          int
	Store              map[string]string
	CommitChan         chan<- CommitEntry
	newCommitReadyChan chan struct{}
}

func (n *Node) Initialize() {
	n.State = Follower
	n.CurrentTerm = 0
	n.VotedFor = nil
	n.Log = []LogEntry{}
	n.CommitIndex = 0
	n.LastApplied = 0
	n.NextIndex = make(map[string]int)
	n.MatchIndex = make(map[string]int)
	n.Store = make(map[string]string)
	n.resetElectionTimer()
	go n.commitChanSender()
}

func (n *Node) resetElectionTimer() {
	if n.State == Leader {
		return // Do not reset election timer if node is the leader
	}

	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}
	timeout := randomElectionTimeout()
	n.ElectionTimer = time.AfterFunc(timeout, func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		if n.State != Leader {
			log.Printf("Election timeout reached, starting new election")
			n.startElection()
		}
	})
	log.Printf("Election timer reset to %v", timeout)
}

func randomElectionTimeout() time.Duration {
	return time.Duration(5+rand.Intn(5)) * time.Second // 5 to 10 seconds
}

func (n *Node) startElection() {
	n.State = Candidate
	n.CurrentTerm++
	n.VotedFor = &n.Id
	n.VoteCount = 0 // Vote for self is done on request vote RPC
	log.Printf("Starting election for term %d", n.CurrentTerm)
	n.resetElectionTimer()
	n.sendRequestVoteRPCs()
}

func (n *Node) sendRequestVoteRPCs() {
	for _, peer := range n.Peers {
		go n.sendRequestVote(peer)
	}
}

func (n *Node) SendAppendEntriesRPCs() {
	for _, peer := range n.Peers {
		if peer == n.Id {
			continue
		}

		log.Printf("Sending AppendEntries RPC to %s", peer)
		go n.sendAppendEntries(peer)
	}
}

func (n *Node) StartHeartbeat() {
	n.HeartbeatTimer = time.NewTicker(1 * time.Second)
	go func() {
		for range n.HeartbeatTimer.C {
			n.SendAppendEntriesRPCs()
		}
	}()
}

func (n *Node) sendRequestVote(peer string) {
	lastLogIndex := len(n.Log) - 1
	lastLogTerm := 0
	if len(n.Log) > 0 {
		lastLogTerm = n.Log[lastLogIndex].Term
	}

	args := RequestVoteArgs{
		Term:         n.CurrentTerm,
		CandidateId:  n.Id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}
	var reply RequestVoteReply
	if call(peer, "Node.RequestVote", args, &reply) {
		n.handleRequestVoteReply(reply)
	}
}

func (n *Node) sendAppendEntries(peer string) {
	n.mu.Lock()

	prevLogIndex := n.NextIndex[peer] - 1
	prevLogTerm := 0
	if prevLogIndex >= 0 && prevLogIndex < len(n.Log) {
		prevLogTerm = n.Log[prevLogIndex].Term
	}

	args := AppendEntriesArgs{
		Term:         n.CurrentTerm,
		LeaderId:     n.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      n.Log[n.NextIndex[peer]:],
		LeaderCommit: n.CommitIndex,
	}
	defer n.mu.Unlock()
	n.dlog("sending AppendEntries to %v: ni=%d, args=%+v", peer, n.NextIndex[peer], args)
	var reply AppendEntriesReply
	if call(peer, "Node.AppendEntries", args, &reply) {
		n.handleAppendEntriesReply(peer, reply)
	}
}

func (n *Node) handleRequestVoteReply(reply RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.dlog("received RequestVoteReply %+v", reply)

	if reply.Term > n.CurrentTerm {
		log.Printf("Received higher term %d, updating current term from %d to %d and converting to follower", reply.Term, n.CurrentTerm, reply.Term)
		n.CurrentTerm = reply.Term
		n.VotedFor = nil
		n.State = Follower
		n.resetElectionTimer()
		return
	}

	if n.State != Candidate {
		n.dlog("while waiting for reply, state = %v", n.State)
		return
	}

	if reply.VoteGranted {
		n.VoteCount++
		log.Printf("Received vote, total votes: %d", n.VoteCount)
		if n.VoteCount > len(n.Peers)/2 {
			log.Printf("Won election for term %d with %d votes", n.CurrentTerm, n.VoteCount)
			n.State = Leader
			n.initializeLeaderState()
			n.StartHeartbeat() // Start sending heartbeats as a leader
		}
	}
}

func (n *Node) initializeLeaderState() {
	n.NextIndex = make(map[string]int)
	n.MatchIndex = make(map[string]int)
	for _, peer := range n.Peers {
		n.NextIndex[peer] = len(n.Log)
		n.MatchIndex[peer] = 0
	}
}

func (n *Node) handleAppendEntriesReply(peer string, reply AppendEntriesReply) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if reply.Term > n.CurrentTerm {
		log.Printf("Received higher term %d, updating current term from %d to %d and converting to follower", reply.Term, n.CurrentTerm, reply.Term)
		n.CurrentTerm = reply.Term
		n.VotedFor = nil
		n.State = Follower
		n.resetElectionTimer()
		return
	}

	if n.State != Leader {
		return
	}

	if reply.Success {
		log.Printf("AppendEntries successful for %s", peer)
		n.MatchIndex[peer] = n.NextIndex[peer]
		n.NextIndex[peer] = len(n.Log)

		savedCommitIndex := n.CommitIndex
		for i := n.CommitIndex + 1; i < len(n.Log); i++ {
			if n.Log[i].Term == n.CurrentTerm {
				matchCount := 1
				for _, peer := range n.Peers {
					if n.MatchIndex[peer] >= i {
						matchCount++
					}
				}
				if matchCount*2 > len(n.Peers)+1 {
					n.CommitIndex = i
				}
			}
		}
		if n.CommitIndex != savedCommitIndex {
			n.dlog("leader sets commitIndex := %d", n.CommitIndex)
			n.newCommitReadyChan <- struct{}{}
		}
	} else {
		n.NextIndex[peer]--
		log.Printf("AppendEntries not successful for %s", peer)
	}
}

func (n *Node) commitChanSender() {
	for range n.newCommitReadyChan {
		n.mu.Lock()
		savedTerm := n.CurrentTerm
		savedLastApplied := n.LastApplied
		var entries []LogEntry
		if n.CommitIndex > n.LastApplied {
			entries = n.Log[n.LastApplied+1 : n.CommitIndex+1]
			n.LastApplied = n.CommitIndex
		}
		n.mu.Unlock()
		n.dlog("commitChanSender entries=%v, savedLastApplied=%d", entries, savedLastApplied)
		for i, entry := range entries {
			n.CommitChan <- CommitEntry{
				Command: entry.Command,
				Index:   savedLastApplied + i + 1,
				Term:    savedTerm,
			}
		}
	}
	n.dlog("commitChanSender done")
}
