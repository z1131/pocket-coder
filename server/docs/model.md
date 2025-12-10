# 数据模型模块 (Model)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

数据模型层定义了与数据库表对应的 Go 结构体（类似 Java 中的 Entity）。

## ER 图

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│      users      │       │    desktops     │       │    sessions     │
├─────────────────┤       ├─────────────────┤       ├─────────────────┤
│ id (PK)         │       │ id (PK)         │       │ id (PK)         │
│ username        │──┐    │ user_id (FK)    │──┐    │ desktop_id (FK) │
│ password_hash   │  │    │ name            │  │    │ agent_type      │
│ email           │  └───►│ device_token    │  └───►│ working_dir     │
│ avatar          │       │ type            │       │ status          │
│ status          │       │ agent_type      │       │ created_at      │
│ created_at      │       │ status          │       └────────┬────────┘
│ updated_at      │       │ last_heartbeat  │                │
└─────────────────┘       │ created_at      │                │
                          └─────────────────┘                │
                                                             │
                          ┌─────────────────┐                │
                          │    messages     │                │
                          ├─────────────────┤                │
                          │ id (PK)         │                │
                          │ session_id (FK) │◄───────────────┘
                          │ role            │
                          │ content         │
                          │ created_at      │
                          └─────────────────┘
```

## 模型定义

### User (用户)

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键 |
| Username | string | 用户名，唯一 |
| PasswordHash | string | 密码哈希 |
| Email | *string | 邮箱，可为空 |
| Avatar | *string | 头像URL，可为空 |
| Status | int8 | 状态：1正常，0禁用 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

### Desktop (电脑设备)

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键 |
| UserID | int64 | 所属用户ID |
| Name | string | 设备名称 |
| DeviceToken | string | 设备唯一标识 |
| Type | string | 类型：local/cloud |
| AgentType | string | AI工具类型 |
| WorkingDir | *string | 工作目录 |
| OSInfo | *string | 操作系统信息 |
| Status | string | 状态：online/offline/busy |
| LastHeartbeat | *time.Time | 最后心跳时间 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

### Session (会话)

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键 |
| DesktopID | int64 | 所属设备ID |
| AgentType | string | AI工具类型 |
| WorkingDir | *string | 工作目录 |
| Status | string | 状态：active/ended |
| StartedAt | time.Time | 开始时间 |
| EndedAt | *time.Time | 结束时间 |
| CreatedAt | time.Time | 创建时间 |

### Message (消息)

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键 |
| SessionID | int64 | 所属会话ID |
| Role | string | 角色：user/assistant/system |
| Content | string | 消息内容 |
| CreatedAt | time.Time | 创建时间 |

## 文件路径

- `internal/model/user.go`
- `internal/model/desktop.go`
- `internal/model/session.go`
- `internal/model/message.go`

## GORM 标签说明

```go
type User struct {
    ID           int64   `gorm:"primaryKey"`           // 主键
    Username     string  `gorm:"size:50;uniqueIndex"`  // 长度50，唯一索引
    PasswordHash string  `gorm:"size:255"`             // 长度255
    Email        *string `gorm:"size:100;uniqueIndex"` // 可为空，唯一索引
}
```

## 注意事项

1. 所有模型都使用 int64 作为主键类型
2. 可为空的字段使用指针类型 (*string, *time.Time)
3. GORM 会自动处理 CreatedAt 和 UpdatedAt
