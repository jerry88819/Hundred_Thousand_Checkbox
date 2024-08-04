package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jerry88819/check-box/redis"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[*websocket.Conn]bool)
	mu       sync.Mutex
)

type Message struct {
	Type       string `json:"type"`
	Index      int    `json:"index,omitempty"`
	Value      bool   `json:"value,omitempty"`
	Data       []bool `json:"data,omitempty"`
	TotalUsers int    `json:"total_users,omitempty"`
} // Message()

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

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

	defer func() {
		mu.Lock()
		delete(clients, ws)
		mu.Unlock()
		ws.Close()
	}()

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
		} // if()

		if msg.Type == "toggle" {
			err := redis.SaveStateToRedis(msg.Index, msg.Value)
			if err != nil {
				log.Println("Error saving state to Redis:", err)
				continue
			} // if()
			broadcast(msg)
		} else if msg.Type == "request_full_state" {
			state, err := redis.GetStateFromRedis()
			if err != nil {
				log.Println("Error getting state from Redis:", err)
				continue
			} // if()
			fullStateMsg := Message{
				Type: "full_state",
				Data: state,
			}
			ws.WriteJSON(fullStateMsg)
		} // else if()
	} // for()
} // handleConnections()

func broadcast(msg Message) {
	mu.Lock()
	defer mu.Unlock()
	log.Println("broad cast!!!")
	ch := make(chan struct{}, 10)
	defer close(ch)
	for client := range clients {
		ch <- struct{}{}
		go func(client *websocket.Conn) {
			defer func() { <-ch }()
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				client.Close()
				delete(clients, client)
			} // if()
		}(client)
	} // for()
} // broadcast()

func broadcastTotalUsers() {
	for {
		time.Sleep(5 * time.Second)
		totalUsers := len(clients)
		userCountMsg := Message{
			Type:       "user_count",
			TotalUsers: totalUsers,
		}
		broadcast(userCountMsg)
	}
}

func setupRoutes() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
} // setupRoutes()

func main() {
	log.Println("WebSocket server started at :8080")

	rdb := redis.Init()
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	} // if()
	fmt.Println(pong)

	setupRoutes()

	go broadcastTotalUsers()

	// 這邊是每 30 秒去推播目前 REDIS 裡面的資料去同步全部人看到的訊息
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
		} // for()
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
} // main()
