package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("USAGE: go run ./server/ <ip_address> <port> [leader_ip_address] [leader_port]")
		return
	}

	ip := os.Args[1]
	port := os.Args[2]

	var leaderIP string
	var leaderPort string
	isLeader := len(os.Args) < 5

	if !isLeader {
		leaderIP = os.Args[3]
		leaderPort = os.Args[4]
	}

	address := ip + ":" + port
	nodeID := ip + ":" + port // Use the combination of IP and port as ID

	// Create and initialize the node
	node := Node{
		Id:         nodeID,
		Peers:      []string{},
		Store:      make(map[string]string),
		LeaderAdd:  leaderIP,
		LeaderPort: leaderPort,
	}
	node.Initialize()

	// Register the node as an RPC server
	err := rpc.Register(&node)
	if err != nil {
		log.Fatalf("Failed to register RPC: %v", err)
	}
	rpc.HandleHTTP()

	// Listen for incoming connections
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}
	// defer listener.Close()

	log.Printf("Node listening on %s", address)

	// If this node is the leader, start sending heartbeats
	if isLeader {
		log.Println("This node is the leader")
		node.State = Leader
		node.Peers = append(node.Peers, nodeID) // Leader adds itself to its peer list
		node.StartHeartbeat()
	} else {
		log.Println("This node is a follower")

		client, err := rpc.DialHTTP("tcp", leaderIP+":"+leaderPort)
		if err != nil {
			log.Fatalf("Failed to connect to leader: %v", err)
		}
		defer client.Close()

		// Send a JoinRPC to the leader
		args := JoinArgs{Id: nodeID}
		var reply JoinReply

		err = client.Call("Node.Join", args, &reply)
		if err != nil {
			log.Fatalf("JoinRPC failed: %v", err)
		}

		// Get the updated list of peers
		node.mu.Lock()
		node.Peers = reply.Peers
		node.mu.Unlock()
		log.Printf("Updated peer list: %v", node.Peers)
	}

	// Start serving RPC requests
	http.Serve(listener, nil)
}
