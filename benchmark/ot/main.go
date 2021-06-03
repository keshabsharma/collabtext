package main

import (
	s "collabtext/server"
	t "collabtext/transform"
	"collabtext/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func simpleOtTest() bool {
	// Test on some simple concurrent OT transforms
	// Check if server and clients have same doc content after OT transforms.

	addr := "localhost:40425"
	room := "doc1"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/" + room}
	numClients := 3
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	//str := "abcd"
	numOps := 6

	numServers := 2

	// servers
	// run servers

	addresses := utils.GetrandomAddresses(numServers)
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

				if err != nil {
					log.Println(err)
					return
				}
			}
		}(i)
	}

	// clients send data
	sender, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < numOps; i++ {
		time.Sleep(time.Millisecond * 300)
		ch := string(str[i])
		c := rand.Intn(numClients)
		ot := t.Operation{1, "insert", 0, ch, "sender" + string(rune(c)), room, ""}
		err = sender.WriteJSON(ot)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	// check documents here

	return false
}

type Response struct {
	Document string `json:"document"`
	Revision uint64 `json:"revision"`
}

func getDocument() string {

	client := &http.Client{}
	req, err := http.NewRequest("GET", "localhost:40425/document/doc1", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject Response
	json.Unmarshal(bodyBytes, &responseObject)
	return responseObject.Document

}

func runServer() {
	numServers := 3

	// servers
	// run servers
	addresses := utils.GetrandomAddresses(numServers)
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
}

func main() {

}
