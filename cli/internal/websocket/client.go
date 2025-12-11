// Package websocket 处理与服务器的 WebSocket 连接
package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 消息类型常量
const (
	TypeHeartbeat = "heartbeat"
	TypePong      = "pong"
	TypePing      = "ping"

	// 终端消息类型
	TypeTerminalInput   = "terminal:input"   // 手机端输入
	TypeTerminalOutput  = "terminal:output"  // 终端输出
	TypeTerminalResize  = "terminal:resize"  // 调整大小
	TypeTerminalExit    = "terminal:exit"    // 终端退出
	TypeTerminalHistory = "terminal:history" // 请求/发送终端历史

	// 服务端 -> 客户端
	TypeSessionCreate = "session:create" // 创建新会话

	// 旧类型（兼容）
	TypeCommand  = "command"  // 来自手机的指令
	TypeAck      = "ack"      // 确认收到
	TypeStream   = "stream"   // 流式输出
	TypeComplete = "complete" // 完成
	TypeError    = "error"    // 错误
	TypeStop     = "stop"     // 停止
)

// Message WebSocket 消息结构
type Message struct {
	Type      string      `json:"type"`
	Content   string      `json:"content,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	MessageID string      `json:"message_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Client WebSocket 客户端
type Client struct {
	conn      *websocket.Conn
	serverURL string
	token     string
	desktopID string
	sendChan  chan []byte
	done      chan struct{}
	mu        sync.Mutex
	isRunning bool
	onMessage func(*Message) // 消息回调
	onConnect func()         // 连接成功回调
	onClose   func()         // 连接关闭回调
}

// NewClient 创建 WebSocket 客户端
// serverURL: HTTP 服务器地址（如 http://localhost:8080）
// token: 访问令牌
// desktopID: 桌面设备 ID
func NewClient(serverURL, token, desktopID string) *Client {
	// 将 HTTP URL 转换为 WebSocket URL
	wsURL := strings.Replace(serverURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = fmt.Sprintf("%s/ws/desktop?token=%s&desktop_id=%s", wsURL, token, desktopID)

	return &Client{
		serverURL: wsURL,
		token:     token,
		desktopID: desktopID,
		sendChan:  make(chan []byte, 256),
		done:      make(chan struct{}),
	}
}

// OnMessage 设置消息回调
func (c *Client) OnMessage(handler func(*Message)) {
	c.onMessage = handler
}

// OnConnect 设置连接成功回调
func (c *Client) OnConnect(handler func()) {
	c.onConnect = handler
}

// OnClose 设置连接关闭回调
func (c *Client) OnClose(handler func()) {
	c.onClose = handler
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.isRunning {
		c.mu.Unlock()
		return fmt.Errorf("客户端已在运行")
	}
	c.mu.Unlock()

	// 建立 WebSocket 连接
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isRunning = true
	c.done = make(chan struct{})
	c.mu.Unlock()

	// 连接成功回调
	if c.onConnect != nil {
		c.onConnect()
	}

	// 启动读写协程
	go c.readPump()
	go c.writePump()

	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return
	}

	c.isRunning = false
	close(c.done)

	if c.conn != nil {
		// 发送关闭帧
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}

	if c.onClose != nil {
		c.onClose()
	}
}

// SendMessage 发送消息
func (c *Client) SendMessage(msg *Message) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.sendChan <- data:
		return nil
	case <-c.done:
		return fmt.Errorf("连接已关闭")
	default:
		return fmt.Errorf("发送缓冲区已满")
	}
}

// readPump 读取消息
func (c *Client) readPump() {
	defer c.Disconnect()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] 读取错误: %v", err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[WS] 解析消息失败: %v", err)
			continue
		}

		// 处理消息
		if c.onMessage != nil {
			c.onMessage(&msg)
		}
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second) // 心跳间隔
	defer func() {
		ticker.Stop()
		c.Disconnect()
	}()

	for {
		select {
		case <-c.done:
			return

		case data := <-c.sendChan:
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[WS] 发送消息失败: %v", err)
				return
			}

		case <-ticker.C:
			// 发送心跳
			heartbeat := &Message{
				Type:      TypeHeartbeat,
				Timestamp: time.Now().UnixMilli(),
			}
			data, _ := json.Marshal(heartbeat)
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[WS] 发送心跳失败: %v", err)
				return
			}
		}
	}
}

// IsRunning 检查是否正在运行
func (c *Client) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}
