package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Error connecting:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error reading:", err)
				return
			}
			fmt.Printf("Received from server: %s\n", message)
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter message: ")
			message, _ := reader.ReadString('\n')
			message = message[:len(message)-1] // Remove newline character
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("Error writing:", err)
				return
			}
		case <-interrupt:
			fmt.Println("Interrupted by user")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error closing connection:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
