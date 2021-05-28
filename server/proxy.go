package server

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func RunKeeper(addr string, serversAddrs []string, ready chan bool) error {
	flag.Parse()
	log.Println("Running keeper ", addr)
	room := newKeeper(serversAddrs)
	initialize(room)
	go room.run()

	router := mux.NewRouter()
	router.HandleFunc("/ws/{room}", room.handleWs)
	http.Handle("/", router)

	//ready <- true
	log.Println("Keeper is Running")
	return http.ListenAndServe(addr, nil)
}
