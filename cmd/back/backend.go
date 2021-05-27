package main

import (
	s "collabtext/server"

	"log"
)

func main() {
	// run servers
	addresses := GetrandomAddresses(4)

	run := func(addr string, ready chan bool) {
		e := s.RunServer(addr, ready)
		if e != nil {
			log.Panicln(e)
		}
	}

	ready1 := make(chan bool)
	ready2 := make(chan bool)
	ready3 := make(chan bool)

	go run(addresses[0], ready1)
	go run(addresses[1], ready2)
	go run(addresses[2], ready3)

	r := <-ready1 && <-ready2 && <-ready3
	if !r {
		log.Fatalln("ded")
	}

	// run keeper
	ready := make(chan bool)
	err := s.RunKeeper(":7777", addresses[:3], ready)
	if err != nil {
		log.Fatalln("keeper did not run")
	}
	if <-ready {
		log.Fatalln("keeper ready")
	}
	log.Println("eee")
}
