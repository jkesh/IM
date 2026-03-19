package controller

import (
	"IM/internal/logic"
	"IM/internal/service"
	"IM/internal/storage/cache"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *service.Hub, w http.ResponseWriter, r *http.Request) {
	tokenString := extractTokenFromRequest(r)
	if tokenString == "" {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}

	userID, err := logic.ValidateToken(tokenString)
	if err != nil {
		log.Printf("token validation failed: %v", err)
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &service.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		ID:   userID,
	}

	client.Hub.Register <- client
	go client.Write()
	go client.Read()
	if cache.Available() {
		go deliverOfflineMessages(client)
	}
}

func deliverOfflineMessages(client *service.Client) {
	offlineKey := "offline:" + client.ID
	msgs, err := cache.RDB.LRange(cache.Ctx, offlineKey, 0, -1).Result()
	if err != nil {
		log.Printf("load offline messages failed for user %s: %v", client.ID, err)
		return
	}

	for _, msg := range msgs {
		select {
		case client.Send <- []byte(msg):
		default:
			log.Printf("offline queue is full for user %s", client.ID)
			return
		}
	}

	if err := cache.RDB.Del(cache.Ctx, offlineKey).Err(); err != nil {
		log.Printf("delete offline messages failed for user %s: %v", client.ID, err)
	}
}
