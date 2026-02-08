package controller

import (
	"IM/internal/service"
	"IM/internal/storage/cache"
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
	tokenString := r.URL.Query().Get("token")
	claims, err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID := claims.UserID
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

	// --- 添加：拉取离线消息 ---
	go func(c *service.Client) {
		offlineKey := "offline:" + c.ID
		// 一次性取出所有消息（或者分页取出）
		msgs, _ := cache.RDB.LRange(cache.Ctx, offlineKey, 0, -1).Result()
		for _, m := range msgs {
			c.Send <- []byte(m)
		}
		// 推送完后删除该 Key
		cache.RDB.Del(cache.Ctx, offlineKey)
	}(client)

	// start
	go client.Write()
	go client.Read()
}
