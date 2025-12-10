// Package model 定义了与数据库表对应的数据结构
package model

import (
	"time"
)

// MessageRole 消息角色常量
const (
	MessageRoleUser      = "user"      // 用户消息
	MessageRoleAssistant = "assistant" // AI 助手响应
	MessageRoleSystem    = "system"    // 系统消息
)

// Message 消息模型
// 对应数据库表 messages
// 存储会话中的每一条消息
type Message struct {
	// ID 消息唯一标识，自增主键
	ID int64 `gorm:"primaryKey" json:"id"`

	// SessionID 所属会话ID，外键关联 sessions.id
	SessionID int64 `gorm:"index;not null" json:"session_id"`

	// Role 消息角色
	// user: 用户发送的消息
	// assistant: AI 助手的响应
	// system: 系统消息（如错误提示）
	Role string `gorm:"size:20;not null" json:"role"`

	// Content 消息内容
	// 使用 TEXT 类型存储，可以存储较长的内容
	Content string `gorm:"type:text;not null" json:"content"`

	// CreatedAt 消息创建时间
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Session 所属会话（多对一关系）
	Session *Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
