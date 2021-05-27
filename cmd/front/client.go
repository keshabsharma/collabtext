package main

import (
	"log"
	"net/url"
	"os"

	t "collabtext/transform"

	"github.com/gorilla/websocket"
)

func main() {
	name := os.Args[1]
	addr := os.Args[2]
	room := "doc1"

	log.Printf("joining room: %s", room)
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/" + room}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var ot t.Operation
			err := c.ReadJSON(&ot)
			log.Println("reading from socket ", ot)
			if err != nil {
				log.Println(err)
				return
			}

		}
	}()

	ot := t.Operation{1, "insert", 1, "a", name, room, ""}
	log.Println("writing to socket", ot)
	err = c.WriteJSON(ot)
	if err != nil {
		log.Println(err)
		return
	}
	select {}

}
