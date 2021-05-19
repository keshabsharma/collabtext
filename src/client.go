package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	// socket connection
	conn *websocket.Conn

	// channel to send message
	messageToSend chan []byte

	room *Room
	name string
}

type Message struct {
	message []byte
	name    string
}

func newClient(name string, room *Room, conn *websocket.Conn) *Client {
	return &Client{
		name:          name,
		conn:          conn,
		messageToSend: make(chan []byte),
		room:          room,
	}
}

func (client *Client) read() {
	defer func() {
		client.conn.Close()
	}()

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		client.room.messageQueue <- Message{msg, client.name}
	}
}

func (client *Client) write() {
	defer func() {
		client.conn.Close()
	}()
	for msg := range client.messageToSend {
		err := client.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
