package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

func main() {
	// parsing arguments
	if len(os.Args) < 3 {
		fmt.Println("ERROR: Invalid arguments")
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

	// creating the node instance
	var nodeType NodeType
	if isLeader {
		nodeType = LEADER
	} else {
		nodeType = FOLLOWER
	}

	node := &Node{
		Address:       Address{IPAddress: ip, Port: port},
		Type:          nodeType,
		ElectionTerm:  0,
		AddressList:   []Address{},
		LeaderAddress: Address{IPAddress: leaderIP, Port: leaderPort},
		Application:   map[string]string{},
	}

	if isLeader {
		node.InitializeLeader()
	} else {
		client, err := rpc.Dial("tcp", leaderIP+":"+leaderPort)
		if err != nil {
			log.Fatal("Error dialing leader:", err)
		}
		defer client.Close()

		var reply string
		err = client.Call("RPCService.AddFollower", Address{IPAddress: ip, Port: port}, &reply)
		if err != nil {
			log.Fatal("Error adding follower to leader:", err)
		}
		fmt.Println(reply)
	}	

	// creating the rpc instance
	rpcService := &RPCService{Node: node}
	rpc.Register(rpcService)

	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Listen error:", err)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("Accept error:", err)
			}
			go rpc.ServeConn(conn)
		}
	}()

	// Print if leader
	if isLeader {
		fmt.Println("This node is a leader")
	} else {
		fmt.Println("This node is a follower")
	}
	fmt.Println("Server listening on " + ip + ":" + port)
	select {}
}
