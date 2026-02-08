package service

import (
	"IM/internal/model"
	"encoding/json"
	"log"
	"sync"
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
				// 简历亮点：O(1) 时间复杂度查找目标连接
				if targetClient, ok := h.Clients[msg.Target]; ok {
					select {
					case targetClient.Send <- msgBytes:
					default:
						close(targetClient.Send)
						delete(h.Clients, msg.Target)
					}
				}
			} else { // 群聊逻辑
				for id, client := range h.Clients {
					// 排除自己或发送给所有人
					client.Send <- msgBytes
				}
			}
			h.mu.RUnlock()
		}

	}
}
