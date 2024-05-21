package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
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
	reader := bufio.NewReader(os.Stdin)

repl:
	for {
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)
		inputArray := strings.Split(input, " ")

		var reply string

		switch inputArray[0] {
		case "exit":
			fmt.Println("Exiting...")
			break repl
		case "ping":
			err = client.Call("RPCService.Ping", struct{}{}, &reply)
		case "set":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of argument")
				break
			}
			err = client.Call("RPCService.Set", inputArray[1:], &reply)
		case "get":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of argument")
				break
			}
			err = client.Call("RPCService.Get", inputArray[1], &reply)
		case "strln":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of argument")
				break
			}
			err = client.Call("RPCService.Strln", inputArray[1], &reply)
		case "del":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of argument")
				break
			}
			err = client.Call("RPCService.Del", inputArray[1], &reply)
		case "append":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of argument")
				break
			}
			err = client.Call("RPCService.Append", inputArray[1:], &reply)
		default:
			fmt.Println("Invalid command")
		}

		if err != nil {
			log.Fatal("Call error:", err)
		}

		fmt.Println(reply)
	}
}
