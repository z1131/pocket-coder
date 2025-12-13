// Package websocket 提供 WebSocket 通信功能
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ClientType 客户端类型
type ClientType int

const (
	ClientTypeMobile  ClientType = iota // 手机端
	ClientTypeDesktop                    // 电脑端
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	hub        *Hub            // 所属的 Hub
	conn       *websocket.Conn // WebSocket 连接
	send       chan []byte     // 发送消息的通道
	clientType ClientType      // 客户端类型
	userID     int64           // 用户ID
	desktopID  int64           // 设备ID（仅电脑端有值）
	processID  string          // 进程ID（仅电脑端有值）
	mu         sync.Mutex      // 保护写操作的互斥锁
}

// 连接配置常量
const (
	// 写超时时间
	writeWait = 10 * time.Second

	// 等待 Pong 响应的超时时间
	pongWait = 60 * time.Second

	// 发送 Ping 的间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 消息最大大小（1MB）
	maxMessageSize = 1024 * 1024
)

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, clientType ClientType, userID, desktopID int64, processID string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256), // 缓冲区大小
		clientType: clientType,
		userID:     userID,
		desktopID:  desktopID,
		processID:  processID,
	}
}

// ReadPump 读取 WebSocket 消息的 goroutine
// 每个客户端连接启动一个 ReadPump
// 负责从 WebSocket 读取消息并分发到 Hub
func (c *Client) ReadPump() {
	// 确保退出时清理资源
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// 设置读取限制
	c.conn.SetReadLimit(maxMessageSize)

	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	// 设置 Pong 处理函数
	// 每次收到 Pong，重置读取超时
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 循环读取消息
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			// 检查是否是正常关闭
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// 处理消息
		c.handleMessage(&msg)
	}
}

// WritePump 写入 WebSocket 消息的 goroutine
// 每个客户端连接启动一个 WritePump
// 负责从 send 通道读取消息并写入 WebSocket
func (c *Client) WritePump() {
	// 创建 Ping 定时器
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 设置写超时
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// send 通道已关闭
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 获取 Writer
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// 写入消息
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 发送 Ping
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 向客户端发送消息
func (c *Client) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 非阻塞发送
	select {
	case c.send <- data:
		return nil
	default:
		// 如果通道已满，说明客户端处理不过来
		log.Printf("Client send buffer full, dropping message")
		return nil
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case TypeHeartbeat:
		// 处理心跳
		c.hub.handleHeartbeat(c)

		// 回复 Pong
		c.SendMessage(NewMessage(TypePong, nil))

	case TypeTerminalInput, TypeTerminalResize:
		// 手机端 → 电脑端：终端输入/调整大小
		if c.clientType == ClientTypeMobile {
			c.hub.handleTerminalToDesktop(c, msg)
		}

	case TypeTerminalOutput, TypeTerminalExit:
		// 电脑端 → 手机端：终端输出/退出
		if c.clientType == ClientTypeDesktop {
			c.hub.handleTerminalToMobile(c, msg)
		}

	case TypeTerminalHistory:
		// 手机端请求历史：从 Redis 获取并返回
		if c.clientType == ClientTypeMobile {
			c.hub.handleTerminalHistoryRequest(c, msg)
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭 send 通道
	select {
	case <-c.send:
		// 通道已关闭
	default:
		close(c.send)
	}
}
