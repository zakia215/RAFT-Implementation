package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"

	"github.com/fatih/color"
)

type ExecuteArgs struct {
	Command string
	Key     string
	Value   string
}

type ExecuteReply struct {
	Response string
}

type LogEntry struct {
	Term    int
	Command string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("ERROR: Invalid arguments")
		fmt.Println("USAGE: go run ./client/ <ip_address> <port>")
		return
	}

	var ip string = os.Args[1]
	var port string = os.Args[2]

	client, err := rpc.DialHTTP("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Connected to", ip+":"+port)

repl:
	for {
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		inputArray := strings.Split(input, " ")

		var reply ExecuteReply
		var err error
		args := ExecuteArgs{
			Command: inputArray[0],
		}

		switch inputArray[0] {
		case "exit":
			fmt.Println("Exiting...")
			break repl
		case "ping":
			args.Key = ""
			args.Value = ""
		case "set":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			args.Key = inputArray[1]
			args.Value = inputArray[2]
		case "get":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			args.Key = inputArray[1]
		case "strln":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			args.Key = inputArray[1]
		case "del":
			if len(inputArray) < 2 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			args.Key = inputArray[1]
		case "append":
			if len(inputArray) < 3 {
				fmt.Println("Invalid number of arguments")
				continue
			}
			args.Key = inputArray[1]
			args.Value = inputArray[2]
		case "log":
			var logReply []LogEntry
			err = client.Call("Node.GetLog", struct{}{}, &logReply)
			if err == nil {
				fmt.Println(logReply)
				for _, entry := range logReply {
					termColor := color.New(color.FgGreen).SprintFunc()
					var commandColor = handleGetCommandColor(fmt.Sprintf("%v", entry.Command))
					fmt.Printf("Term: %s%s | Command: %s\n", termColor(entry.Term), color.New(color.Reset).SprintFunc()(), commandColor(entry.Command))
				}
				continue
			}
		default:
			fmt.Println("Invalid command")
			continue
		}

		if args.Command == "log" {
			continue
		}

		err = client.Call("Node.Execute", args, &reply)
		if err != nil {
			log.Fatal("Call error:", err)
		}

		fmt.Println(reply.Response)
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
