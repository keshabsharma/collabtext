package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 1024
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type Room struct {

	// connected clients
	clients map[string]map[*Client]bool

	// will hold message that will be send to all clients
	messageQueue chan Message

	connect    chan *Client
	disconnect chan *Client
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.connect:
			log.Println("connection received ")
			clients := r.clients[c.name]
			if clients == nil {
				r.clients[c.name] = make(map[*Client]bool)
			}
			r.clients[c.name][c] = true
		case c := <-r.disconnect:
			clients := r.clients[c.name]
			if clients != nil {
				delete(clients, c)
				c.conn.Close()
				close(c.messageToSend)
			}
		case m := <-r.messageQueue:
			clients := r.clients[m.name]

			// send message to all clients
			for client := range clients {
				client.messageToSend <- m.message
			}

		}
	}
}

func (r *Room) handleWs(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	roomname, ok := vars["room"]
	log.Println("got connection for room: ", roomname)
	if !ok {
		log.Println("missing room name")
	}

	// upgrade request to websocket connection
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer c.Close()

	client := newClient(roomname, r, c)

	r.connect <- client
	defer func() { r.disconnect <- client }()
	go client.write()
	client.read()

}

func newRoom() *Room {
	return &Room{
		clients:      make(map[string]map[*Client]bool),
		messageQueue: make(chan Message),
		connect:      make(chan *Client),
		disconnect:   make(chan *Client),
	}
}
