package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var CountClient = 0
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool { // throw an error if the number of users is more than 2
		if CountClient >= 2 {
			return false
		}
		CountClient++
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // connections with bool are active or not?
var clientsLock sync.Mutex

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	clientsLock.Lock()
	clients[conn] = true
	clientsLock.Unlock()

	defer func() {
		clientsLock.Lock()
		delete(clients, conn)
		clientsLock.Unlock()
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		fmt.Printf("Received message: %s\n", p)

		// Get sender
		sender := conn

		clientsLock.Lock()
		for client := range clients {
			if client != sender {
				err = client.WriteMessage(messageType, p) // from the server to the client send message
				if err != nil {
					fmt.Println("Error writing:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
		clientsLock.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws", handleConnection)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil) // port number: 8080
}
