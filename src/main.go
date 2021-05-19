package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var addr = flag.String("addr", ":7777", "http service address")

func main() {
	flag.Parse()
	log.Println("Starting server ")
	room := newRoom()
	go room.run()

	router := mux.NewRouter()
	router.HandleFunc("/ws/{room}", room.handleWs)

	http.Handle("/", router)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
