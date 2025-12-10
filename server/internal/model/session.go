// Package model 定义了与数据库表对应的数据结构
package model

import (
	"time"
)

// SessionStatus 会话状态常量
const (
	SessionStatusActive = "active" // 活跃中
	SessionStatusEnded  = "ended"  // 已结束
)

// Session 会话模型
// 对应数据库表 sessions
// 表示用户与 AI 的一次对话会话
// 一个设备可以有多个会话，类似于聊天应用中的对话窗口
type Session struct {
	// ID 会话唯一标识，自增主键
	ID int64 `gorm:"primaryKey" json:"id"`

	// DesktopID 所属设备ID，外键关联 desktops.id
	DesktopID int64 `gorm:"index;not null" json:"desktop_id"`

	// AgentType 使用的 AI 工具类型
	// 记录创建会话时使用的工具，因为设备可能切换工具
	AgentType string `gorm:"size:50;not null" json:"agent_type"`

	// WorkingDir 会话的工作目录
	// 创建会话时的项目路径
	WorkingDir *string `gorm:"size:500" json:"working_dir,omitempty"`

	// Status 会话状态
	// active: 活跃中，可以继续对话
	// ended: 已结束
	Status string `gorm:"size:20;default:active;index" json:"status"`

	// StartedAt 会话开始时间
	StartedAt time.Time `gorm:"autoCreateTime" json:"started_at"`

	// EndedAt 会话结束时间
	// 仅当状态为 ended 时有值
	EndedAt *time.Time `json:"ended_at,omitempty"`

	// CreatedAt 创建时间（与 StartedAt 相同）
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Desktop 所属设备（多对一关系）
	Desktop *Desktop `gorm:"foreignKey:DesktopID" json:"desktop,omitempty"`

	// Messages 会话中的所有消息（一对多关系）
	Messages []Message `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}
