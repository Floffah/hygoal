package main

import (
	"fmt"
	"hygoal/internal/network"
)

func main() {
	fmt.Println("hello world")

	err := network.StartQuicServer()
	if err != nil {
		panic(err)
	}
}
