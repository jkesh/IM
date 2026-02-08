package service

import (
	"IM/internal/model"
	"IM/internal/storage/cache"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Hub struct {
	//已注册的客户端
	Clients map[string]*Client
	//广播消息
	Broadcast chan []byte
	//注册请求通道
	Register chan *Client
	//注销通道
	Unregister chan *Client
	//加锁
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()

			// 在 hub.go 的 Run() 函数中修改 Broadcast 分支
		case msgBytes := <-h.Broadcast:
			var msg model.Message
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Println("消息解析失败:", err)
				continue
			}

			h.mu.RLock()
			if msg.Type == 1 { // 私聊
				if targetClient, ok := h.Clients[msg.Target]; ok {
					targetClient.Send <- msgBytes
				} else {
					// --- 核心亮点：存入 Redis 离线列表 ---
					// 使用 RPush 将消息存入名为 "offline:用户ID" 的 List 中
					cache.RDB.RPush(cache.Ctx, "offline:"+msg.Target, msgBytes)
					// 设置过期时间（如 7 天），防止 Redis 内存被僵尸用户撑爆
					cache.RDB.Expire(cache.Ctx, "offline:"+msg.Target, time.Hour*24*7)
				}
			} else { // 群聊逻辑
				for _, client := range h.Clients {
					// 排除自己或发送给所有人
					client.Send <- msgBytes
				}
			}
			h.mu.RUnlock()
		}

	}
}
