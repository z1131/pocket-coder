// Package model 定义了与数据库表对应的数据结构
package model

import (
	"time"
)

// Message 消息模型
// 对应数据库表 messages
// 表示会话中的一条具体消息，例如用户输入或 AI 回复
type Message struct {
	// ID 消息唯一标识，自增主键
	ID int64 `gorm:"primaryKey" json:"id"`

	// SessionID 所属会话ID，外键关联 sessions.id
	SessionID int64 `gorm:"index;not null" json:"session_id"`

	// Role 消息角色，例如 "user", "assistant", "system"
	Role string `gorm:"size:50;not null" json:"role"`

	// Content 消息内容
	Content string `gorm:"type:text;not null" json:"content"`

	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}