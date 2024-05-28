package main

import (
	"bufio"
	"common"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"

	"github.com/fatih/color"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("ERROR: Invalid arguments")
		fmt.Println("USAGE: go run ./client/ <ip_address> <port>")
		return
	}

	var ip string = os.Args[1]
	var port string = os.Args[2]

	client, err := rpc.Dial("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	reader := bufio.NewReader(os.Stdin)

repl:
	for {
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)
		inputArray := strings.Split(input, " ")

		var reply interface{}
		var err error

		switch inputArray[0] {
		case "exit":
			fmt.Println("Exiting...")
			break repl
		case "ping":
			reply = new(string)
			err = client.Call("RPCService.Ping", struct{}{}, reply.(*string))
		case "set":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			reply = new(string)
			err = client.Call("RPCService.Set", inputArray[1:], reply.(*string))
		case "get":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			reply = new(string)
			err = client.Call("RPCService.Get", inputArray[1], reply.(*string))
		case "strln":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			reply = new(string)
			err = client.Call("RPCService.Strln", inputArray[1], reply.(*string))
		case "del":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			reply = new(string)
			err = client.Call("RPCService.Del", inputArray[1], reply.(*string))
		case "append":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			reply = new(string)
			err = client.Call("RPCService.Append", inputArray[1:], reply.(*string))
		case "log":
			reply = new([]common.LogEntry)
			err = client.Call("RPCService.GetLog", struct{}{}, reply.(*[]common.LogEntry))
			if err == nil {
				for _, entry := range *reply.(*[]common.LogEntry) {
					termColor := color.New(color.FgGreen).SprintFunc()
					var commandColor = handleGetCommandColor(entry.Command)
					fmt.Printf("Term: %s%s | Command: %s\n", termColor(entry.Term), color.New(color.Reset).SprintFunc()(), commandColor(entry.Command))
				}
				continue
			}
		default:
			fmt.Println("Invalid command")
			continue
		}

		if err != nil {
			log.Fatal("Call error:", err)
		}

		if r, ok := reply.(*string); ok {
			fmt.Println(*r)
		}
	}
}

func handleGetCommandColor(command string) func(a ...interface{}) string {
	if strings.HasPrefix(command, "SET") {
		return color.New(color.FgBlue).SprintFunc()
	} else if strings.HasPrefix(command, "DEL") {
		return color.New(color.FgRed).SprintFunc()
	} else if strings.HasPrefix(command, "APPEND") {
		return color.New(color.FgYellow).SprintFunc()
	}
	return color.New(color.FgHiMagenta).SprintFunc()
}
