// Package model 定义了与数据库表对应的数据结构
// 这些结构体类似于 Java 中的 Entity 类
package model

import (
	"time"
)

// User 用户模型
// 对应数据库表 users
// 存储用户的基本信息，包括认证凭据
type User struct {
	// ID 用户唯一标识，自增主键
	ID int64 `gorm:"primaryKey" json:"id"`

	// Username 用户名，用于登录，全局唯一
	// 长度限制 50 字符，建立唯一索引
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"`

	// PasswordHash 密码的 bcrypt 哈希值
	// 永远不要存储明文密码！
	PasswordHash string `gorm:"size:255;not null" json:"-"` // json:"-" 表示序列化时忽略此字段

	// Email 用户邮箱，可选，用于找回密码等
	// 使用指针类型表示可以为 NULL
	Email *string `gorm:"size:100;uniqueIndex" json:"email,omitempty"`

	// Avatar 用户头像 URL，可选
	Avatar *string `gorm:"size:500" json:"avatar,omitempty"`

	// Status 账号状态
	// 1: 正常
	// 0: 禁用
	Status int8 `gorm:"default:1" json:"status"`

	// CreatedAt 创建时间，由 GORM 自动填充
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// UpdatedAt 更新时间，由 GORM 自动更新
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Desktops 用户拥有的电脑设备（一对多关系）
	// 这是 GORM 的关联关系，不会在数据库中创建字段
	Desktops []Desktop `gorm:"foreignKey:UserID" json:"desktops,omitempty"`
}

// TableName 指定表名
// GORM 会使用这个方法返回的表名，而不是默认的复数形式
func (User) TableName() string {
	return "users"
}
