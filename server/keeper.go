package server

import (
	t "collabtext/transform"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/rpc"
	"sync"
	"time"

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

	serversLock sync.Mutex
	processLock sync.Mutex
}

type ServerStatus struct {
	Addr     string
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
			for client := range clients {
				client.otToSend <- m.ot
			}

		}
	}
}

type DocRes struct {
	Document string `json:"document"`
	Revision uint64 `json:"revision"`
}

func (r *Keeper) handleDoc(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	doc, ok := vars["doc"]
	if !ok {
		w.WriteHeader(400)
		return
	}

	if _, ok = r.servers[doc]; !ok {
		w.WriteHeader(400)
		return
	}

	// get document from server
	i := getLatestServerIndex(r.servers[doc])
	res, err := GetDocument(r.servers[doc][i].Addr, doc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func (r *Keeper) handleWs(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	roomname, ok := vars["room"]
	log.Println("got connection for room: ", roomname)
	if !ok {
		log.Println("missing room name")
	}

	// upgrade request to websocket connection
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
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
	r.processLock.Lock()
	defer r.processLock.Unlock()

	i := getLatestServerIndex(servers)
	server := servers[i]
	processedOt, err := r.processTransformation(server.Addr, ot)

	if err != nil {
		return nil, err
	}

	r.UpdateServerRevision(ot.Document, server.Addr, processedOt.Revision)

	// replicate the processed transforms
	go r.broadcastTransformation(servers, i, processedOt)

	return processedOt, nil
}

func (r *Keeper) broadcastTransformation(servers []*ServerStatus, processingServer int, ot *t.Operation) {
	for i, v := range servers {
		if i == processingServer {
			continue
		}

		// only send to synced servers. separate syncing in background
		if ot.Revision-v.Revision > 1 {
			continue
		}

		go func(s *ServerStatus) {
			var rev *uint64
			rev, err := applyTransformation(s.Addr, ot)
			if err == nil {
				r.UpdateServerRevision(ot.Document, s.Addr, *rev)
			}

			r.printInfo()

		}(v)
	}
}

func (r *Keeper) processTransformation(addr string, ot *t.Operation) (*t.Operation, error) {
	var ret *t.Operation
	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	e = conn.Call("Server.ProcessTransformation", ot, &ret)
	if e != nil {
		conn.Close()
		return nil, e
	}
	return ret, conn.Close()
}

func applyTransformation(addr string, ot *t.Operation) (*uint64, error) {
	req := make([]t.Operation, 0)
	req = append(req, *ot)
	return applyTransformations(addr, ot.Document, req)
}

////// Sync

func (r *Keeper) runSync() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		r.syncDocuments()
	}

}

func (r *Keeper) syncDocuments() {
	for k, d := range r.servers {
		r.syncDocument(k, d)
	}
}

func (r *Keeper) syncDocument(document string, servers []*ServerStatus) {
	maxRev := uint64(0)
	for _, v := range servers {
		if v.Revision >= maxRev {
			maxRev = v.Revision
		}
	}
	serversToUpdate := make([]*ServerStatus, 0)
	lowRev := uint64(math.MaxUint64)
	latestServer := ""

	for _, v := range servers {
		if v.Revision < maxRev {
			serversToUpdate = append(serversToUpdate, v)
			if v.Revision <= lowRev {
				lowRev = v.Revision
			}
		}
		if v.Revision == maxRev {
			latestServer = v.Addr
		}
	}

	if len(serversToUpdate) == 0 {
		return
	}

	//log.Println("primary server ", latestServer, " from rev ", lowRev+1)

	transforms, err := getTransformations(latestServer, document, lowRev+1)

	if err != nil || len(transforms) == 0 {
		return
	}

	for _, v := range serversToUpdate {
		go func(s *ServerStatus) {
			var rev *uint64
			rev, e := applyTransformations(s.Addr, document, transforms[s.Revision-lowRev:])
			if e == nil {
				r.UpdateServerRevision(document, s.Addr, *rev)
			} else {
				log.Println(err)
			}
			r.printInfo()
		}(v)
	}
}

func (r *Keeper) UpdateServerRevision(document string, addr string, rev uint64) {
	r.serversLock.Lock()
	defer r.serversLock.Unlock()

	if servers, ok := r.servers[document]; ok {
		for i, v := range servers {
			if v.Addr == addr {
				r.servers[document][i].Revision = rev
				return
			}
		}
	}

}

func getTransformations(addr string, document string, fromRevision uint64) ([]t.Operation, error) {
	var ret TransformsList
	req := &GetTransforms{document, fromRevision}

	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	ret.T = nil
	e = conn.Call("Server.GetTransformations", req, &ret)
	if e != nil {
		conn.Close()
		return nil, e
	}
	if ret.T == nil {
		ret.T = make([]t.Operation, 0)
	}
	return ret.T, conn.Close()
}

func applyTransformations(addr string, document string, transforms []t.Operation) (*uint64, error) {
	var ret *uint64
	req := new(TransformsList)
	req.Document = document
	req.T = append(req.T, transforms...)
	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	e = conn.Call("Server.ApplyTransformations", req, &ret)
	if e != nil {
		conn.Close()
		return nil, e
	}
	return ret, conn.Close()
}

func GetDocument(addr string, document string) (*DocRes, error) {
	var res *DocRes

	conn, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return nil, e
	}

	e = conn.Call("Server.GetDocument", document, &res)
	if e != nil {
		conn.Close()
		return nil, e
	}

	return res, conn.Close()
}

///// DEBUG

func (r *Keeper) printInfo() {
	log.Println("server statuses")
	for k, v := range r.servers {
		fmt.Println("DOCUMENT: ", k)
		for _, s := range v {
			fmt.Println("Addr: ", s.Addr, " Revision:", s.Revision)
		}
	}
}
