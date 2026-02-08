package controller

import (
	"IM/internal/service"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *service.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// generate new id for client
	userID := r.URL.Query().Get("uid")
	if userID == "" {
		userID = "guest"
	}

	client := &service.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		ID:   userID,
	}

	client.Hub.Register <- client

	// start
	go client.Write()
	go client.Read()
}
