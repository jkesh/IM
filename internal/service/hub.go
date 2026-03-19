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
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	offlineTTL time.Duration
	mu         sync.RWMutex
}

func NewHub(offlineTTL time.Duration) *Hub {
	if offlineTTL <= 0 {
		offlineTTL = 7 * 24 * time.Hour
	}

	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		offlineTTL: offlineTTL,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.addClient(client)
		case client := <-h.Unregister:
			h.removeClient(client)
		case msgBytes := <-h.Broadcast:
			h.dispatch(msgBytes)
		}
	}
}

func (h *Hub) dispatch(msgBytes []byte) {
	var msg model.Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		log.Printf("unmarshal broadcast message failed: %v", err)
		return
	}

	switch msg.Type {
	case model.MessageTypePrivate:
		h.dispatchPrivate(msg.Target, msgBytes)
	case model.MessageTypeGroup:
		for _, client := range h.listClients() {
			h.send(client, msgBytes)
		}
	}
}

func (h *Hub) dispatchPrivate(target string, msgBytes []byte) {
	targetClient, ok := h.getClient(target)
	if ok {
		h.send(targetClient, msgBytes)
		return
	}

	if !cache.Available() {
		return
	}

	offlineKey := "offline:" + target
	if err := cache.RDB.RPush(cache.Ctx, offlineKey, msgBytes).Err(); err != nil {
		log.Printf("push offline message failed: %v", err)
		return
	}
	if err := cache.RDB.Expire(cache.Ctx, offlineKey, h.offlineTTL).Err(); err != nil {
		log.Printf("set offline message ttl failed: %v", err)
	}
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Clients[client.ID] = client
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	registeredClient, ok := h.Clients[client.ID]
	if !ok || registeredClient != client {
		return
	}

	delete(h.Clients, client.ID)
	close(client.Send)
}

func (h *Hub) getClient(userID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.Clients[userID]
	return client, ok
}

func (h *Hub) listClients() []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*Client, 0, len(h.Clients))
	for _, client := range h.Clients {
		clients = append(clients, client)
	}
	return clients
}

func (h *Hub) send(client *Client, msgBytes []byte) {
	select {
	case client.Send <- msgBytes:
	default:
		h.removeClient(client)
		client.Conn.Close()
	}
}
