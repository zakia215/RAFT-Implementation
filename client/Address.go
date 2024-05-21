package main

import "fmt"

type Address struct {
	IPAddress string
	Port      int
}

func (a *Address) PrintAddress() {
	fmt.Printf("Address: %s:%d\n", a.IPAddress, a.Port)
}
