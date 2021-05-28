package server

import (
	t "collabtext/transform"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"sync"

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

type Keeper struct {

	// connected clients
	clients map[string]map[*WsClient]bool

	messageQueue chan Message

	connect    chan *WsClient
	disconnect chan *WsClient

	servers      map[string][]*ServerStatus
	serversAddrs []string
	serversLock  sync.Mutex
}

type ServerStatus struct {
	addr     string
	Revision uint64
}

func (r *Keeper) run() {
	for {
		select {
		case c := <-r.connect:
			log.Println("connection received ")
			clients := r.clients[c.name]
			if clients == nil {
				r.clients[c.name] = make(map[*WsClient]bool)
			}
			r.clients[c.name][c] = true
		case c := <-r.disconnect:
			clients := r.clients[c.name]
			if clients != nil {
				delete(clients, c)
				c.conn.Close()
				close(c.otToSend)
			}
		case m := <-r.messageQueue:
			clients := r.clients[m.name]

			log.Println("sending from msg queue for doc ", m.name)
			for client := range clients {
				client.otToSend <- m.ot
			}

		}
	}
}

func (r *Keeper) handleWs(w http.ResponseWriter, req *http.Request) {
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

func newKeeper(serversAddrs []string) *Keeper {
	return &Keeper{
		clients:      make(map[string]map[*WsClient]bool),
		messageQueue: make(chan Message),
		connect:      make(chan *WsClient),
		disconnect:   make(chan *WsClient),
		servers:      make(map[string][]*ServerStatus),
		serversAddrs: serversAddrs,
	}
}

func initialize(k *Keeper) {

	servers := make([]*ServerStatus, 0)
	for _, v := range k.serversAddrs {
		servers = append(servers, &ServerStatus{v, 0})
	}

	k.servers["doc1"] = servers
}

func (r *Keeper) ProcessOperation(ot *t.Operation) (*t.Operation, error) {
	servers, ok := r.servers[ot.Document]
	if !ok {
		return nil, fmt.Errorf("document does not exist")
	}

	i := getLatestServerIndex(servers)
	server := servers[i]

	processedOt, err := r.processTransformation(server.addr, ot)
	if err != nil {
		return nil, err
	}

	// update local server status
	r.serversLock.Lock()
	defer r.serversLock.Unlock()
	r.servers[ot.Document][i].Revision = processedOt.Revision

	// replicate the processed transforms
	go r.broadcastTransformation(servers, i, processedOt)

	return processedOt, nil
}

func (r *Keeper) broadcastTransformation(servers []*ServerStatus, processingServer int, ot *t.Operation) {
	for i, v := range servers {
		if i == processingServer {
			continue
		}
		// only send to synced servers
		if ot.Revision-v.Revision > 1 {
			continue
		}

		var succ *bool
		succ, _ = applyTransformation(v.addr, ot)
		if *succ {
			r.serversLock.Lock()
			defer r.serversLock.Unlock()
			r.servers[ot.Document][i].Revision = ot.Revision
		}
	}
}

func (r *Keeper) processTransformation(addr string, ot *t.Operation) (*t.Operation, error) {
	var ret *t.Operation
	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	e = conn.Call("Server.ProcessTransformation", ot, ret)
	if e != nil {
		conn.Close()
		return nil, e
	}
	return ret, conn.Close()
}

func applyTransformation(addr string, ot *t.Operation) (*bool, error) {
	var ret *bool
	req := make([]*t.Operation, 1)
	req = append(req, ot)
	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	e = conn.Call("Server.applyTransformations", req, ret)
	if e != nil {
		conn.Close()
		return nil, e
	}
	return ret, conn.Close()
}
