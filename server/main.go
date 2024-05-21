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

	var ip string = os.Args[1]
	var port string = os.Args[2]

	var leaderIP string = ""
	var leaderPort string = ""
	var isLeader bool = false

	if len(os.Args) >= 5 {
		leaderIP = os.Args[3]
		leaderPort = os.Args[4]
		isLeader = true
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

	fmt.Println("Server listening on " + ip + ":" + port)
	select {}
}
