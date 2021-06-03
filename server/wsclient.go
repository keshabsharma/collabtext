package server

import (
	t "collabtext/transform"
	"log"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	// socket connection
	conn *websocket.Conn

	// channel to send message
	//messageToSend chan Message
	otToSend chan t.Operation

	room *Keeper
	name string
}

type Message struct {
	//message []byte
	ot   t.Operation
	name string
}

func newClient(name string, room *Keeper, conn *websocket.Conn) *WsClient {
	return &WsClient{
		name:     name,
		conn:     conn,
		otToSend: make(chan t.Operation),
		room:     room,
	}
}

func (client *WsClient) read() {
	defer func() {
		client.conn.Close()
	}()

	for {
		var ot t.Operation
		err := client.conn.ReadJSON(&ot)
		log.Println(ot)
		if err != nil {
			log.Println("readjson error ", err)
			// connection ended
			return
		}

		newOt, err := client.room.ProcessOperation(&ot)
		if err != nil {
			ot.Error = err.Error()
			newOt = &ot
		}
		client.room.messageQueue <- Message{*newOt, client.name}
	}
}

func (client *WsClient) write() {
	defer func() {
		client.conn.Close()
	}()
	for ot := range client.otToSend {
		err := client.conn.WriteJSON(ot)
		if err != nil {
			return
		}
	}
}
