package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jerry88819/check-box/redis"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[*websocket.Conn]bool)
)

type Message struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Value bool   `json:"value,omitempty"`
	Data  []bool `json:"data,omitempty"`
} // Message()

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	state, err := redis.GetStateFromRedis()
	if err != nil {
		log.Println("Error getting state from Redis:", err)
		return
	}

	fullStateMsg := Message{
		Type: "full_state",
		Data: state,
	}
	ws.WriteJSON(fullStateMsg)

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
                log.Println("WebSocket closed normally:", err)
            } else {
                log.Println("WebSocket error:", err)
            } // else()

			delete(clients, ws)
			break
		}

		if msg.Type == "toggle" {
			err := redis.SaveStateToRedis(msg.Index, msg.Value)
			if err != nil {
				log.Println("Error saving state to Redis:", err)
				continue
			}
			broadcast(msg)
		} else if msg.Type == "request_full_state" {
			state, err := redis.GetStateFromRedis()
			if err != nil {
				log.Println("Error getting state from Redis:", err)
				continue
			}
			fullStateMsg := Message{
				Type: "full_state",
				Data: state,
			}
			ws.WriteJSON(fullStateMsg)
		}
	}
} // handleConnections()

func broadcast(msg Message) {
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
} // broadcast()

func setupRoutes() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
} // setupRoutes()

func main() {
	log.Println("WebSocket server started at :8080")
	rdb := redis.Init()

    // 測試連線是否成功
    pong, err := rdb.Ping(context.Background()).Result()
    if err != nil {
        panic(err)
    }
    fmt.Println(pong) // 輸出 "PONG"

	setupRoutes()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			state, err := redis.GetStateFromRedis()
			if err != nil {
				log.Println("Error getting state from Redis:", err)
				continue
			}
			fullStateMsg := Message{
				Type: "full_state",
				Data: state,
			}
			broadcast(fullStateMsg)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
} // main()
