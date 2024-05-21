package main

import "fmt"

type RPCService struct {
	Node *Node
}

func (r *RPCService) Get(key string, reply *string) error {
	fmt.Println("GET ", key)
	*reply = r.Node.GetValue(key)
	return nil
}

func (r *RPCService) Ping(_ struct{}, reply *string) error {
	fmt.Println("PING")
	*reply = "PONG"
	return nil
}
