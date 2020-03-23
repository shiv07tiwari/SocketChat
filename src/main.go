package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	clients  = make(map[*websocket.Conn]bool) // connected clients
	brodcast = make(chan message)             // brodcase channel
	upgrader = websocket.Upgrader{}
)

type message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handle connection called")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	clients[ws] = true

	for {
		var msg message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error : %v", err)
			delete(clients, ws)
			break
		}
		brodcast <- msg
	}
}

func handleMessages() {
	fmt.Println("Handle message called")
	for {
		msg := <-brodcast

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
