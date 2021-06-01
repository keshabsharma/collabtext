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
	keeper := newKeeper(serversAddrs)
	initialize(keeper)
	go keeper.run()
	go keeper.runSync()

	router := mux.NewRouter()
	router.HandleFunc("/ws/{room}", keeper.handleWs)
	router.HandleFunc("/document/{doc}", keeper.handleDoc)
	http.Handle("/", router)

	//ready <- true
	log.Println("Keeper is Running")
	return http.ListenAndServe(addr, nil)
}
