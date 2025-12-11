// Package model 定义了与数据库表对应的数据结构
package model

import (
	"time"
)

// DesktopType 设备类型常量
const (
	DesktopTypeLocal = "local" // 本地电脑
	DesktopTypeCloud = "cloud" // 云端电脑
)

// DesktopStatus 设备状态常量
const (
	DesktopStatusOnline  = "online"  // 在线
	DesktopStatusOffline = "offline" // 离线
	DesktopStatusBusy    = "busy"    // 忙碌（正在执行任务）
)

// Desktop 电脑设备模型
// 对应数据库表 desktops
// 表示用户绑定的电脑端设备
type Desktop struct {
	// ID 设备唯一标识，自增主键
	ID int64 `gorm:"primaryKey" json:"id"`

	// UserID 所属用户ID，外键关联 users.id
	UserID int64 `gorm:"index;not null" json:"user_id"`

	// Name 设备名称，用于用户识别设备
	// 例如: "MacBook-Home", "Office-PC"
	Name string `gorm:"size:100;not null" json:"name"`

	// DeviceUUID 设备唯一标识（客户端持久化的 UUID）
	// 用于设备去重：同一用户 + 同一 DeviceUUID = 同一台设备
	// 即使用户更改主机名，此 UUID 也不会变化
	DeviceUUID string `gorm:"size:64;index" json:"-"`

	// DeviceToken 设备认证令牌
	// 由服务端生成，用于设备身份验证
	// 全局唯一，建立唯一索引
	DeviceToken string `gorm:"size:64;uniqueIndex;not null" json:"-"` // 不对外暴露

	// Type 设备类型
	// local: 本地安装的 CLI 工具
	// cloud: 云端托管的开发环境（预留）
	Type string `gorm:"size:20;default:local" json:"type"`

	// AgentType 使用的 AI 工具类型
	// 例如: "claude-code", "aider", "goose"
	AgentType string `gorm:"size:50;default:claude-code" json:"agent_type"`

	// WorkingDir 当前工作目录
	// 电脑端报告的当前项目路径
	WorkingDir *string `gorm:"size:500" json:"working_dir,omitempty"`

	// OSInfo 操作系统信息
	// 例如: "macOS 14.0", "Windows 11", "Ubuntu 22.04"
	OSInfo *string `gorm:"size:200" json:"os_info,omitempty"`

	// Status 设备状态
	// online: 在线
	// offline: 离线
	// busy: 正在执行任务
	// 注意: 实时状态从 Redis 获取，此字段用于持久化
	Status string `gorm:"size:20;default:offline;index" json:"status"`

	// LastHeartbeat 最后心跳时间
	// 用于判断设备是否仍然在线
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`

	// CreatedAt 设备绑定时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// UpdatedAt 最后更新时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// User 所属用户（多对一关系）
	// 通过 UserID 字段关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Sessions 设备上的所有会话（一对多关系）
	Sessions []Session `gorm:"foreignKey:DesktopID" json:"sessions,omitempty"`
}

// TableName 指定表名
func (Desktop) TableName() string {
	return "desktops"
}
