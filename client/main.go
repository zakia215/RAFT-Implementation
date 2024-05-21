package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

func main() {
	// parsing arguments
	if len(os.Args) != 3 {
		fmt.Println("ERROR: Invalid arguments")
		fmt.Println("USAGE: go run ./client/ <ip_address> <port>")
		return
	}

	var ip string = os.Args[1]
	var port string = os.Args[2]

	// dialling the server
	client, err := rpc.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Dial error:", err)
	}

	// sending the request
	var reply string
	err = client.Call("RPCService.Ping", struct{}{}, &reply)

	if err != nil {
		log.Fatal("Call error:", err)
	}

	fmt.Println(reply)
}
