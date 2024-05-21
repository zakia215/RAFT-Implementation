package main

import "fmt"

type Address struct {
	IPAddress string
	Port      string
}

func (a *Address) PrintAddress() {
	fmt.Printf("Address: %s:%s\n", a.IPAddress, a.Port)
}
