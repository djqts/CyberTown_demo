package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// OnMessageFunc 用户消息回调，发送原始消息和 UserToken。
type OnMessageFunc func(userToken string, msg []byte)

// Client 代表一个 WebSocket 连接。
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	UserToken string
	onMessage OnMessageFunc
}

// readPump 从 WebSocket 读取消息并调用回调。
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if c.onMessage != nil {
			c.onMessage(c.UserToken, msg)
		}
	}
}

// writePump 将消息写入 WebSocket，并发送心跳 ping。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send 将消息 JSON 编码后发送。
func (c *Client) Send(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	select {
	case c.send <- data:
	default:
	}
	return nil
}
