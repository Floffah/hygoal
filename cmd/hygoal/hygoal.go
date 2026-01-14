package main

import (
	"hygoal/internal/network"
)

func main() {
	err := network.StartQuicServer()
	if err != nil {
		panic(err)
	}
}
