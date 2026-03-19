package service

import (
	"IM/internal/logic"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	Hub  *Hub
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

func (c *Client) Read() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		msg, shouldDispatch, err := logic.NormalizeIncomingMessage(c.ID, payload)
		if err != nil {
			log.Printf("invalid message from user %s: %v", c.ID, err)
			continue
		}
		if !shouldDispatch {
			continue
		}

		if err := logic.SaveMessage(msg); err != nil {
			log.Printf("save message failed for user %s: %v", c.ID, err)
		}

		messageBytes, err := logic.MarshalMessage(msg)
		if err != nil {
			log.Printf("marshal message failed for user %s: %v", c.ID, err)
			continue
		}
		c.Hub.Broadcast <- messageBytes
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
