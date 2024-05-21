package main

import (
	"errors"
	"fmt"
	"strconv"
)

type RPCService struct {
	Node *Node
}

func (r *RPCService) Ping(_ struct{}, reply *string) error {
	fmt.Println("PING")
	*reply = "PONG"
	return nil
}

func (r *RPCService) Get(key string, reply *string) error {
	fmt.Println("GET ", key)
	*reply = r.Node.GetValue(key)
	return nil
}

func (r *RPCService) Set(args []string, reply *string) error {
	if len(args) < 2 {
		return errors.New("insufficient number of arguments")
	}
	key := args[0]
	val := args[1]
	fmt.Println("SET " + key + " " + val)
	r.Node.SetValue(key, val)
	*reply = "OK"
	return nil
}

func (r *RPCService) Strln(key string, reply *string) error {
	fmt.Println("STRLN " + key)
	*reply = strconv.Itoa(len(r.Node.GetValue(key)))
	return nil
}

func (r *RPCService) Del(key string, reply *string) error {
	fmt.Println("DEL " + key)
	*reply = r.Node.GetValue(key)
	delete(r.Node.Application, key)
	return nil
}

func (r *RPCService) Append(args []string, reply *string) error {
	if len(args) < 2 {
		return errors.New("insufficient number of arguments")
	}
	r.Node.SetValue(args[0], r.Node.GetValue(args[0])+args[1])
	*reply = "OK"
	return nil
}
