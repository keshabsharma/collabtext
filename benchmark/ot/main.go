package main

import (
	s "collabtext/server"
	"collabtext/utils"
	"log"
	"time"
)

func simpleOtTest() bool {
	// Test on some simple OT transforms
	// Check if server and clients have same doc content after OT transforms.

	return false
}

func runServer() {
	numServers := 3

	// servers
	// run servers
	addresses := utils.GetrandomAddresses(numServers)
	run := func(addr string, ready chan bool) {
		e := s.RunServer(addr, ready)
		if e != nil {
			log.Panicln(e)
		}
	}

	for i := 0; i < numServers; i++ {
		ready1 := make(chan bool)
		go run(addresses[i], ready1)
		<-ready1
	}
	// run keeper
	ready := make(chan bool)
	go s.RunKeeper(":40425", addresses, ready)

	time.Sleep(time.Second * 2)
}

func main() {

}
