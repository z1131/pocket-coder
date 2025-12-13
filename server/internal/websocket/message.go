// Package websocket 提供 WebSocket 通信功能
// 实现手机端和电脑端的实时双向通信
package websocket

import (
	"time"
)

// MessageType 消息类型常量
const (
	// 电脑端 → 服务端
	TypeHeartbeat      = "heartbeat"       // 心跳
	TypeAgentResponse  = "agent:response"  // AI 完整响应
	TypeAgentStream    = "agent:stream"    // AI 流式输出
	TypeAgentStatus    = "agent:status"    // AI 状态变更

	// 终端消息类型（双向透传）
	TypeTerminalInput   = "terminal:input"   // 手机端 → 电脑端：终端输入
	TypeTerminalOutput  = "terminal:output"  // 电脑端 → 手机端：终端输出
	TypeTerminalResize  = "terminal:resize"  // 手机端 → 电脑端：调整终端大小
	TypeTerminalExit    = "terminal:exit"    // 电脑端 → 手机端：终端退出
	TypeTerminalHistory = "terminal:history" // 双向：请求/返回终端历史

	// 服务端 → 电脑端
	TypeUserMessage    = "user:message"    // 用户发送的消息
	TypeSessionCreate  = "session:create"  // 创建新会话
	TypeSessionClose   = "session:close"   // 关闭会话

	// 服务端 → 手机端
	TypeDesktopOnline  = "desktop:online"  // 电脑上线
	TypeDesktopOffline = "desktop:offline" // 电脑下线

	// 通用
	TypeError          = "error"           // 错误消息
	TypePong           = "pong"            // 心跳响应
)

// Message WebSocket 消息结构
// 所有消息都使用这个统一的结构
type Message struct {
	Type      string      `json:"type"`                 // 消息类型
	Payload   interface{} `json:"payload"`              // 消息内容
	Timestamp int64       `json:"timestamp"`            // 时间戳（毫秒）
	MessageID string      `json:"message_id,omitempty"` // 消息ID，用于追踪
}

// NewMessage 创建新消息
func NewMessage(msgType string, payload interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewMessageWithID 创建带消息ID的新消息
func NewMessageWithID(msgType string, payload interface{}, messageID string) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
		MessageID: messageID,
	}
}

// ==================== Payload 类型定义 ====================

// UserMessagePayload 用户消息 Payload
// 手机端发送消息时使用
type UserMessagePayload struct {
	DesktopID int64  `json:"desktop_id"`          // 目标设备ID
	SessionID int64  `json:"session_id,omitempty"` // 会话ID（可选，不传使用当前活跃会话）
	Content   string `json:"content"`              // 消息内容
}

// AgentResponsePayload AI 响应 Payload
// 电脑端返回 AI 完整响应时使用
type AgentResponsePayload struct {
	SessionID int64  `json:"session_id"` // 会话ID
	Content   string `json:"content"`    // 响应内容
	Role      string `json:"role"`       // 角色（通常是 assistant）
}

// AgentStreamPayload AI 流式输出 Payload
// 电脑端返回 AI 流式输出时使用
type AgentStreamPayload struct {
	SessionID int64  `json:"session_id"` // 会话ID
	Delta     string `json:"delta"`      // 增量内容
}

// AgentStatusPayload AI 状态 Payload
// 电脑端报告 AI 工作状态时使用
type AgentStatusPayload struct {
	Status    string `json:"status"`     // 状态：running / idle
	SessionID int64  `json:"session_id,omitempty"` // 会话ID
}

// DesktopStatusPayload 设备状态 Payload
// 通知手机端设备上线/下线时使用
type DesktopStatusPayload struct {
	DesktopID int64  `json:"desktop_id"` // 设备ID
	Status    string `json:"status,omitempty"` // 状态（可选）
}

// SessionCreatePayload 创建会话 Payload
// 通知电脑端创建新会话时使用
type SessionCreatePayload struct {
	SessionID  int64  `json:"session_id"`            // 会话ID
	WorkingDir string `json:"working_dir,omitempty"` // 工作目录
	IsDefault  bool   `json:"is_default,omitempty"`  // 是否为默认会话（需要本地显示）
}

// SessionClosePayload 关闭会话 Payload
// 通知电脑端关闭会话时使用
type SessionClosePayload struct {
	SessionID int64 `json:"session_id"` // 会话ID
}

// ErrorPayload 错误消息 Payload
type ErrorPayload struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
}

// HeartbeatPayload 心跳 Payload
type HeartbeatPayload struct {
	// 目前心跳不需要额外数据
}
