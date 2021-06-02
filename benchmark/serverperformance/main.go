package main

import (
	"log"
	"net/url"
	"time"

	s "collabtext/server"
	t "collabtext/transform"

	"github.com/gorilla/websocket"
)

func main() {

	addr := "localhost:40425"
	room := "doc1"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/" + room}
	numClients := 20
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//str := "abcd"
	numOps := len(str)
	timers := make(map[string]time.Time)
	times := make([][]time.Duration, numOps)
	for i := 0; i < numOps; i++ {
		times[i] = make([]time.Duration, 0)
	}

	numServers := 1

	// servers
	// run servers
	addresses := GetrandomAddresses(numServers)
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

	// clients
	for i := 0; i < numClients; i++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal(err)
		}

		go func(j int) {
			for {
				var ot t.Operation
				err := c.ReadJSON(&ot)
				elapsed := time.Since(timers[ot.Str])

				if err != nil {
					log.Println(err)
					return
				}
				times[ot.Revision-1] = append(times[ot.Revision-1], elapsed)
			}
		}(i)
	}

	// clients keep sending data
	sender, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < numOps; i++ {
		time.Sleep(time.Millisecond * 100)
		c := string(str[i])
		timers[c] = time.Now()
		ot := t.Operation{1, "insert", 0, c, "sender", room, ""}
		err = sender.WriteJSON(ot)
		if err != nil {
			log.Println(err)
			return
		}
	}

	time.Sleep(time.Second * 5)

	for i, v := range times {
		log.Println("rev", i+1)
		log.Println("clients ", len(v), " avg", avg(v))
	}

}

func avg(times []time.Duration) float64 {
	total := 0
	for _, t := range times {
		total = total + int(t.Milliseconds())
	}
	return float64(total) / float64(len(times))

}
