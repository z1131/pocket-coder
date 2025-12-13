# Remote Claude Code - 最终技术方案

> 手机远程控制电脑 AI 编程助手的完整解决方案

---

## 一、产品概述

### 1.1 产品定位

让程序员通过手机随时随地控制电脑上的 AI 编程工具（Claude Code 等），实现移动办公。

### 1.2 核心功能

| 功能 | 描述 |
|------|------|
| 用户系统 | 注册、登录、账号管理 |
| 设备管理 | 一个账号可绑定多台电脑 |
| 远程控制 | 手机发送指令，电脑执行，结果返回 |
| 会话管理 | 保存历史对话，支持多会话 |
| 实时同步 | WebSocket 双向通信 |

### 1.3 用户流程

```
首次使用：
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  1. 电脑端                        2. 手机端                      │
│     $ remote-claude login            打开网页/App                │
│              │                            │                     │
│              ▼                            ▼                     │
│     显示: "请访问 xxx.com/device     用户登录账号                │
│            输入代码: ABCD-1234"           │                     │
│              │                            │                     │
│              │         3. 用户在手机上输入设备码                 │
│              │◄───────────────────────────┤                     │
│              │                            │                     │
│              ▼                            ▼                     │
│     ✓ 登录成功，开始服务            看到电脑上线                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

日常使用：
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  电脑端（后台运行）                  手机端                      │
│     $ remote-claude start              │                        │
│              │                         │                        │
│              │    1. 选择要控制的电脑  │                        │
│              │◄────────────────────────┤                        │
│              │                         │                        │
│              │    2. 发送: "帮我写一个登录页面"                  │
│              │◄────────────────────────┤                        │
│              │                         │                        │
│      Claude Code 执行                  │                        │
│              │                         │                        │
│              │    3. 返回执行结果       │                        │
│              ├────────────────────────►│                        │
│              │                         │                        │
│                                   查看代码输出                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、系统架构

### 2.1 整体架构图

```
                                    ┌───────────────────────────────────────┐
                                    │               阿里云                   │
                                    │                                       │
                                    │  ┌─────────────┐   ┌─────────────┐   │
                                    │  │   MySQL     │   │    Redis    │   │
                                    │  │   (RDS)     │   │   (阿里云)   │   │
                                    │  │             │   │             │   │
                                    │  │  用户数据   │   │  在线状态   │   │
                                    │  │  设备数据   │   │  授权码     │   │
                                    │  │  会话历史   │   │  会话缓存   │   │
                                    │  │  消息记录   │   │  Pub/Sub   │   │
                                    │  └──────▲──────┘   └──────▲──────┘   │
                                    │         │                 │          │
                                    │         └────────┬────────┘          │
                                    │                  │                   │
┌──────────────────┐                │   ┌──────────────┴──────────────┐    │
│   📱 手机端       │   HTTPS/WSS   │   │        Go 服务端            │    │
│   (React PWA)    │◄──────────────►│   │                             │    │
│                  │                │   │  ┌───────────────────────┐  │    │
│ • 登录/注册      │                │   │  │      HTTP API         │  │    │
│ • 设备列表       │                │   │  │  • /api/auth/*        │  │    │
│ • 对话界面       │                │   │  │  • /api/desktop/*     │  │    │
│ • 实时输出       │                │   │  │  • /api/session/*     │  │    │
└──────────────────┘                │   │  └───────────────────────┘  │    │
                                    │   │  ┌───────────────────────┐  │    │
                                    │   │  │    WebSocket Hub      │  │    │
                                    │   │  │  • 连接管理            │  │    │
                                    │   │  │  • 消息路由            │  │    │
                                    │   │  │  • 状态同步            │  │    │
                                    │   │  └───────────────────────┘  │    │
                                    │   └──────────────▲──────────────┘    │
                                    │                  │                   │
                                    └──────────────────┼───────────────────┘
                                                       │
                           ┌───────────────────────────┼───────────────────────────┐
                           │                           │                           │
                           ▼                           ▼                           ▼
                  ┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
                  │  💻 电脑端 1     │         │  💻 电脑端 2     │         │  💻 电脑端 N     │
                  │  (Go CLI)       │         │  (Go CLI)       │         │  (Go CLI)       │
                  │                 │         │                 │         │                 │
                  │  ┌───────────┐  │         │  ┌───────────┐  │         │  ┌───────────┐  │
                  │  │ AgentAPI  │  │         │  │ AgentAPI  │  │         │  │ AgentAPI  │  │
                  │  └─────┬─────┘  │         │  └─────┬─────┘  │         │  └─────┬─────┘  │
                  │        │        │         │        │        │         │        │        │
                  │  ┌─────▼─────┐  │         │  ┌─────▼─────┐  │         │  ┌─────▼─────┐  │
                  │  │  Claude   │  │         │  │   Aider   │  │         │  │  Goose    │  │
                  │  │   Code    │  │         │  │           │  │         │  │           │  │
                  │  └───────────┘  │         │  └───────────┘  │         │  └───────────┘  │
                  └─────────────────┘         └─────────────────┘         └─────────────────┘
                      家里电脑                    公司电脑                     云服务器
```

### 2.2 技术栈

| 组件 | 技术选型 | 版本 | 说明 |
|------|---------|------|------|
| **服务端** | Go + Gin + GORM | Go 1.21+ | 高性能、易部署 |
| **数据库** | MySQL | 8.0 | 阿里云 RDS，持久化存储 |
| **缓存** | Redis | 7.0 | 阿里云 Redis，状态管理 |
| **电脑端** | Go CLI | Go 1.21+ | 单二进制分发 |
| **AI 集成** | AgentAPI | latest | Claude Code 等 |
| **手机端** | React + TypeScript | React 18 | PWA 支持 |
| **通信协议** | WebSocket + REST | - | 实时 + 请求响应 |

### 2.3 为什么选择这些技术

| 选择 | 理由 |
|------|------|
| **Go** | 单二进制部署、低内存、原生并发支持、适合长连接场景 |
| **Gin** | Go 最流行的 Web 框架，性能好，文档全 |
| **GORM** | Go 最主流的 ORM，类似 JPA，学习成本低 |
| **MySQL** | 你已有阿里云 RDS，持久化核心业务数据 |
| **Redis** | 你已有阿里云 Redis，处理实时状态和缓存 |
| **AgentAPI** | 成熟的 Claude Code HTTP 封装，避免重复造轮子 |
| **React PWA** | 可安装到手机桌面，体验接近原生 App |

### 2.4 数据存储职责划分

| 数据类型 | MySQL | Redis | 说明 |
|---------|:-----:|:-----:|------|
| 用户账号 | ✓ | - | 持久化存储 |
| 设备信息 | ✓ | ✓ | MySQL 持久化，Redis 存在线状态 |
| 会话记录 | ✓ | ✓ | MySQL 持久化，Redis 缓存热点数据 |
| 消息内容 | ✓ | - | 持久化存储 |
| 设备授权码 | - | ✓ | 15分钟 TTL 自动过期 |
| 在线状态 | - | ✓ | 实时状态，Set 数据结构 |
| JWT 黑名单 | - | ✓ | Token 过期时间 TTL |
| 跨实例消息 | - | ✓ | Pub/Sub 广播 |

---

## 三、数据库设计

### 3.1 MySQL 表结构

#### ER 图

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

#### 建表 SQL

```sql
-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    email VARCHAR(100) UNIQUE COMMENT '邮箱',
    avatar VARCHAR(500) COMMENT '头像URL',
    status TINYINT DEFAULT 1 COMMENT '1:正常 0:禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 电脑端设备表
CREATE TABLE desktops (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '所属用户',
    name VARCHAR(100) NOT NULL COMMENT '设备名称',
    device_token VARCHAR(64) NOT NULL UNIQUE COMMENT '设备唯一标识',
    type ENUM('local', 'cloud') DEFAULT 'local' COMMENT '设备类型',
    agent_type VARCHAR(50) DEFAULT 'claude-code' COMMENT 'AI工具类型',
    working_dir VARCHAR(500) COMMENT '工作目录',
    os_info VARCHAR(200) COMMENT '操作系统信息',
    status ENUM('online', 'offline', 'busy') DEFAULT 'offline' COMMENT '在线状态',
    last_heartbeat DATETIME COMMENT '最后心跳时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_device_token (device_token),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='电脑端设备表';

-- 会话表
CREATE TABLE sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    desktop_id BIGINT NOT NULL COMMENT '所属设备',
    agent_type VARCHAR(50) NOT NULL COMMENT 'AI工具类型',
    working_dir VARCHAR(500) COMMENT '工作目录',
    status ENUM('active', 'ended') DEFAULT 'active' COMMENT '会话状态',
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    ended_at DATETIME COMMENT '结束时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (desktop_id) REFERENCES desktops(id) ON DELETE CASCADE,
    INDEX idx_desktop_id (desktop_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

-- 消息表
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL COMMENT '所属会话',
    role ENUM('user', 'assistant', 'system') NOT NULL COMMENT '消息角色',
    content TEXT NOT NULL COMMENT '消息内容',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    INDEX idx_session_id (session_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';
```

### 3.2 Redis 数据结构设计

```yaml
# ==================== 设备授权码 ====================
# 用于电脑端登录流程，15分钟自动过期

# 设备码 -> 设备信息
device_code:{code}:
  type: Hash
  ttl: 15 minutes
  fields:
    device_token: "abc123..."
    user_code: "ABCD-1234"
    status: "pending" | "authorized"
    user_id: "1"  # 授权后写入

# 用户码 -> 设备码（反向索引）
user_code:{user_code}:
  type: String
  ttl: 15 minutes
  value: "{device_code}"

# ==================== 在线状态 ====================

# 在线设备集合
online:desktops:
  type: Set
  value: [1, 2, 3]  # desktop_id 列表

# 设备心跳时间
desktop:{id}:heartbeat:
  type: String
  ttl: 2 minutes  # 超时则认为离线
  value: "1704067200"  # Unix 时间戳

# 用户的在线设备（方便查询某用户有哪些设备在线）
user:{id}:online_desktops:
  type: Set
  value: [1, 3]  # desktop_id 列表

# ==================== 会话缓存 ====================

# 活跃会话缓存（热点数据）
session:{id}:cache:
  type: Hash
  ttl: 1 hour
  fields:
    desktop_id: "1"
    agent_type: "claude-code"
    working_dir: "/path/to/project"
    status: "active"

# 设备的当前活跃会话
desktop:{id}:active_session:
  type: String
  value: "123"  # session_id

# ==================== JWT 黑名单 ====================

# 已登出的 Token（用于强制失效）
jwt:blacklist:{token_hash}:
  type: String
  ttl: {token 剩余有效期}
  value: "1"

# ==================== Pub/Sub 频道 ====================

# 用户消息频道（多实例部署时广播）
channel: user:{user_id}:messages

# 设备状态变更频道
channel: desktop:status
```

### 3.3 Redis 操作示例

```go
// internal/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
    return &RedisCache{client: client}
}

// ==================== 设备授权码 ====================

type DeviceCodeInfo struct {
    DeviceToken string `json:"device_token"`
    UserCode    string `json:"user_code"`
    Status      string `json:"status"`
    UserID      int64  `json:"user_id,omitempty"`
}

// 创建设备授权码
func (c *RedisCache) CreateDeviceCode(ctx context.Context, code string, info *DeviceCodeInfo) error {
    key := fmt.Sprintf("device_code:%s", code)
    data, _ := json.Marshal(info)
    
    pipe := c.client.Pipeline()
    pipe.Set(ctx, key, data, 15*time.Minute)
    pipe.Set(ctx, fmt.Sprintf("user_code:%s", info.UserCode), code, 15*time.Minute)
    _, err := pipe.Exec(ctx)
    return err
}

// 获取设备授权码信息
func (c *RedisCache) GetDeviceCode(ctx context.Context, code string) (*DeviceCodeInfo, error) {
    key := fmt.Sprintf("device_code:%s", code)
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    var info DeviceCodeInfo
    json.Unmarshal(data, &info)
    return &info, nil
}

// 通过用户码获取设备码
func (c *RedisCache) GetDeviceCodeByUserCode(ctx context.Context, userCode string) (string, error) {
    return c.client.Get(ctx, fmt.Sprintf("user_code:%s", userCode)).Result()
}

// 授权设备码
func (c *RedisCache) AuthorizeDeviceCode(ctx context.Context, code string, userID int64) error {
    key := fmt.Sprintf("device_code:%s", code)
    info, err := c.GetDeviceCode(ctx, code)
    if err != nil || info == nil {
        return fmt.Errorf("device code not found")
    }
    
    info.Status = "authorized"
    info.UserID = userID
    data, _ := json.Marshal(info)
    
    // 保留原有 TTL
    ttl, _ := c.client.TTL(ctx, key).Result()
    return c.client.Set(ctx, key, data, ttl).Err()
}

// ==================== 在线状态 ====================

// 设置设备在线
func (c *RedisCache) SetDesktopOnline(ctx context.Context, desktopID, userID int64) error {
    pipe := c.client.Pipeline()
    pipe.SAdd(ctx, "online:desktops", desktopID)
    pipe.SAdd(ctx, fmt.Sprintf("user:%d:online_desktops", userID), desktopID)
    pipe.Set(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID), time.Now().Unix(), 2*time.Minute)
    _, err := pipe.Exec(ctx)
    return err
}

// 设置设备离线
func (c *RedisCache) SetDesktopOffline(ctx context.Context, desktopID, userID int64) error {
    pipe := c.client.Pipeline()
    pipe.SRem(ctx, "online:desktops", desktopID)
    pipe.SRem(ctx, fmt.Sprintf("user:%d:online_desktops", userID), desktopID)
    pipe.Del(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID))
    _, err := pipe.Exec(ctx)
    return err
}

// 更新心跳
func (c *RedisCache) UpdateHeartbeat(ctx context.Context, desktopID int64) error {
    return c.client.Set(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID), time.Now().Unix(), 2*time.Minute).Err()
}

// 检查设备是否在线
func (c *RedisCache) IsDesktopOnline(ctx context.Context, desktopID int64) bool {
    return c.client.SIsMember(ctx, "online:desktops", desktopID).Val()
}

// 获取用户的在线设备列表
func (c *RedisCache) GetUserOnlineDesktops(ctx context.Context, userID int64) ([]int64, error) {
    result, err := c.client.SMembers(ctx, fmt.Sprintf("user:%d:online_desktops", userID)).Result()
    if err != nil {
        return nil, err
    }
    
    ids := make([]int64, 0, len(result))
    for _, s := range result {
        var id int64
        fmt.Sscanf(s, "%d", &id)
        ids = append(ids, id)
    }
    return ids, nil
}

// ==================== 会话缓存 ====================

// 设置当前活跃会话
func (c *RedisCache) SetActiveSession(ctx context.Context, desktopID, sessionID int64) error {
    return c.client.Set(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID), sessionID, 0).Err()
}

// 获取当前活跃会话
func (c *RedisCache) GetActiveSession(ctx context.Context, desktopID int64) (int64, error) {
    result, err := c.client.Get(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID)).Int64()
    if err == redis.Nil {
        return 0, nil
    }
    return result, err
}

// ==================== JWT 黑名单 ====================

// 将 Token 加入黑名单
func (c *RedisCache) BlacklistToken(ctx context.Context, tokenHash string, expireAt time.Time) error {
    ttl := time.Until(expireAt)
    if ttl <= 0 {
        return nil // Token 已过期，无需加入黑名单
    }
    return c.client.Set(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenHash), "1", ttl).Err()
}

// 检查 Token 是否在黑名单中
func (c *RedisCache) IsTokenBlacklisted(ctx context.Context, tokenHash string) bool {
    return c.client.Exists(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenHash)).Val() > 0
}

// ==================== Pub/Sub ====================

// 发布用户消息（多实例广播）
func (c *RedisCache) PublishUserMessage(ctx context.Context, userID int64, message interface{}) error {
    data, _ := json.Marshal(message)
    return c.client.Publish(ctx, fmt.Sprintf("user:%d:messages", userID), data).Err()
}

// 订阅用户消息
func (c *RedisCache) SubscribeUserMessages(ctx context.Context, userID int64) *redis.PubSub {
    return c.client.Subscribe(ctx, fmt.Sprintf("user:%d:messages", userID))
}
```

---

## 四、API 设计

### 4.1 接口总览

```yaml
# ==================== 认证模块 ====================
POST   /api/auth/register              # 用户注册
POST   /api/auth/login                 # 用户登录
POST   /api/auth/logout                # 用户登出（Token 加入黑名单）
POST   /api/auth/refresh               # 刷新 Token
POST   /api/auth/device/code           # [电脑端] 获取设备授权码
GET    /api/auth/device/status         # [电脑端] 轮询授权状态
POST   /api/auth/device/authorize      # [手机端] 授权设备

# ==================== 用户模块 ====================
GET    /api/user/profile               # 获取当前用户信息
PUT    /api/user/profile               # 更新用户信息

# ==================== 设备模块 ====================
GET    /api/desktops                   # 获取我的电脑列表（含在线状态）
GET    /api/desktops/:id               # 获取电脑详情
PUT    /api/desktops/:id               # 更新电脑信息（名称等）
DELETE /api/desktops/:id               # 删除电脑

# ==================== 会话模块 ====================
GET    /api/desktops/:id/sessions      # 获取某电脑的会话列表
POST   /api/desktops/:id/sessions      # 创建新会话
GET    /api/sessions/:id               # 获取会话详情（含消息历史）
DELETE /api/sessions/:id               # 删除会话

# ==================== WebSocket ====================
WS     /ws/desktop                     # 电脑端 WebSocket 连接
WS     /ws/mobile                      # 手机端 WebSocket 连接
```

### 4.2 核心接口详情

#### 4.2.1 用户注册

```yaml
POST /api/auth/register
Content-Type: application/json

Request:
{
    "username": "zhangsan",
    "password": "123456",
    "email": "zhangsan@example.com"      # 可选
}

Response 200:
{
    "code": 0,
    "message": "注册成功",
    "data": {
        "user_id": 1,
        "username": "zhangsan"
    }
}

Response 400:
{
    "code": 1001,
    "message": "用户名已存在"
}
```

#### 4.2.2 用户登录

```yaml
POST /api/auth/login
Content-Type: application/json

Request:
{
    "username": "zhangsan",
    "password": "123456"
}

Response 200:
{
    "code": 0,
    "message": "登录成功",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 86400,
        "user": {
            "id": 1,
            "username": "zhangsan",
            "email": "zhangsan@example.com",
            "avatar": null
        }
    }
}
```

#### 4.2.3 用户登出

```yaml
POST /api/auth/logout
Authorization: Bearer <token>

Response 200:
{
    "code": 0,
    "message": "登出成功"
}

# 后端处理：将当前 Token 加入 Redis 黑名单
```

#### 4.2.4 设备获取授权码（电脑端登录流程）

```yaml
POST /api/auth/device/code
Content-Type: application/json

Request:
{
    "device_token": "abc123...",        # 设备唯一标识，首次为空则生成
    "device_name": "MacBook-Home",      # 设备名称
    "os_info": "macOS 14.0"             # 操作系统信息
}

Response 200:
{
    "code": 0,
    "data": {
        "device_code": "GH7s9dKJh2...",         # 内部使用的长码
        "user_code": "ABCD-1234",               # 用户输入的短码
        "verification_uri": "https://xxx.com/device",
        "expires_in": 900,                       # 15分钟过期
        "interval": 5                            # 轮询间隔（秒）
    }
}

# 后端处理：将授权码信息存入 Redis，15分钟 TTL
```

#### 4.2.5 电脑端轮询授权状态

```yaml
GET /api/auth/device/status?device_code=GH7s9dKJh2...

Response 202 (等待授权):
{
    "code": 0,
    "data": {
        "status": "pending"
    }
}

Response 200 (授权成功):
{
    "code": 0,
    "data": {
        "status": "authorized",
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "desktop_id": 1
    }
}

Response 410 (已过期):
{
    "code": 1002,
    "message": "授权码已过期"
}

# 后端处理：从 Redis 读取授权码状态
```

#### 4.2.6 获取电脑列表

```yaml
GET /api/desktops
Authorization: Bearer <token>

Response 200:
{
    "code": 0,
    "data": {
        "desktops": [
            {
                "id": 1,
                "name": "MacBook-Home",
                "type": "local",
                "agent_type": "claude-code",
                "status": "online",           # 从 Redis 获取实时状态
                "working_dir": "/Users/zhang/projects/myapp",
                "os_info": "macOS 14.0",
                "last_heartbeat": "2024-01-15T10:30:00Z"
            },
            {
                "id": 2,
                "name": "Office-PC",
                "type": "local",
                "agent_type": "claude-code",
                "status": "offline",
                "working_dir": "D:\\projects",
                "os_info": "Windows 11",
                "last_heartbeat": "2024-01-14T18:00:00Z"
            }
        ]
    }
}

# 后端处理：
# 1. 从 MySQL 获取用户的设备列表
# 2. 从 Redis 获取每个设备的在线状态
# 3. 合并返回
```

### 4.3 WebSocket 消息协议

#### 消息格式

```typescript
interface WSMessage {
    type: string;              // 消息类型
    payload: object;           // 消息内容
    timestamp: number;         // 时间戳（毫秒）
    message_id?: string;       // 消息ID（用于追踪）
}
```

#### 消息类型

```yaml
# ========== 电脑端 → 服务端 ==========
heartbeat:              # 心跳（更新 Redis 状态）
  payload: {}

agent:response:         # AI 完整响应
  payload:
    session_id: 1
    content: "好的，我来帮你创建登录页面..."
    role: "assistant"

agent:stream:           # AI 流式输出（打字机效果）
  payload:
    session_id: 1
    delta: "好的，"      # 增量内容
    
agent:status:           # AI 状态变更
  payload:
    status: "running" | "idle"

# ========== 服务端 → 电脑端 ==========
user:message:           # 用户发送的消息
  payload:
    session_id: 1
    content: "帮我写一个登录页面"

session:create:         # 创建新会话
  payload:
    working_dir: "/path/to/project"

# ========== 服务端 → 手机端 ==========
desktop:online:         # 电脑上线通知
  payload:
    desktop_id: 1

desktop:offline:        # 电脑下线通知
  payload:
    desktop_id: 1

agent:response:         # AI 响应（转发）
agent:stream:           # AI 流式输出（转发）
agent:status:           # AI 状态（转发）

# ========== 手机端 → 服务端 ==========
user:message:           # 发送消息给指定电脑
  payload:
    desktop_id: 1
    session_id: 1       # 可选，不传则使用当前活跃会话
    content: "帮我写一个登录页面"
```

---

## 五、项目结构

### 5.1 服务端结构

```
server/
├── cmd/
│   └── server/
│       └── main.go                 # 入口文件
│
├── internal/
│   ├── config/
│   │   └── config.go               # 配置加载
│   │
│   ├── model/                      # 数据模型 (对应 Entity)
│   │   ├── user.go
│   │   ├── desktop.go
│   │   ├── session.go
│   │   └── message.go
│   │
│   ├── repository/                 # 数据访问层 (对应 Mapper/DAO)
│   │   ├── user_repo.go
│   │   ├── desktop_repo.go
│   │   ├── session_repo.go
│   │   └── message_repo.go
│   │
│   ├── cache/                      # Redis 缓存层
│   │   └── redis.go
│   │
│   ├── service/                    # 业务逻辑层
│   │   ├── user_service.go
│   │   ├── auth_service.go
│   │   ├── desktop_service.go
│   │   └── session_service.go
│   │
│   ├── handler/                    # HTTP 处理器 (对应 Controller)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── desktop_handler.go
│   │   └── session_handler.go
│   │
│   ├── websocket/                  # WebSocket 模块
│   │   ├── hub.go                  # 连接管理中心
│   │   ├── client.go               # 客户端连接
│   │   ├── message.go              # 消息定义
│   │   └── handler.go              # 消息处理
│   │
│   └── middleware/                 # 中间件
│       ├── auth.go                 # JWT 认证
│       ├── cors.go                 # 跨域
│       └── logger.go               # 日志
│
├── pkg/                            # 公共工具包
│   ├── jwt/
│   │   └── jwt.go
│   ├── response/
│   │   └── response.go             # 统一响应格式
│   └── util/
│       └── util.go
│
├── configs/
│   ├── config.yaml                 # 配置模板
│   └── config.prod.yaml            # 生产配置
│
├── scripts/
│   └── migrate.sql                 # 数据库迁移脚本
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

### 5.2 电脑端 CLI 结构

```
desktop/
├── cmd/
│   └── remote-claude/
│       └── main.go                 # CLI 入口
│
├── internal/
│   ├── agent/
│   │   ├── agent.go                # Agent 核心逻辑
│   │   ├── agentapi.go             # AgentAPI 集成
│   │   └── websocket.go            # WebSocket 客户端
│   │
│   ├── auth/
│   │   └── device_auth.go          # 设备登录流程
│   │
│   └── config/
│       └── config.go               # 本地配置管理
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 5.3 手机端结构

```
mobile/
├── public/
│   ├── index.html
│   └── manifest.json               # PWA 配置
│
├── src/
│   ├── api/                        # API 调用
│   │   ├── auth.ts
│   │   ├── desktop.ts
│   │   └── session.ts
│   │
│   ├── components/                 # 组件
│   │   ├── MessageList.tsx
│   │   ├── MessageInput.tsx
│   │   ├── DesktopList.tsx
│   │   └── StatusBar.tsx
│   │
│   ├── hooks/                      # 自定义 Hooks
│   │   ├── useWebSocket.ts
│   │   └── useAuth.ts
│   │
│   ├── pages/                      # 页面
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Home.tsx                # 设备列表
│   │   ├── Chat.tsx                # 对话界面
│   │   └── DeviceAuth.tsx          # 设备授权页
│   │
│   ├── stores/                     # 状态管理
│   │   ├── authStore.ts
│   │   └── chatStore.ts
│   │
│   ├── App.tsx
│   └── main.tsx
│
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── tsconfig.json
```

---

## 六、核心流程

### 6.1 电脑端登录流程（Device Flow）

```
┌──────────────┐                    ┌──────────────┐                    ┌──────────────┐
│   电脑端     │                    │    服务端    │                    │  浏览器/手机  │
│   (CLI)      │                    │              │                    │              │
└──────┬───────┘                    └──────┬───────┘                    └──────┬───────┘
       │                                   │                                   │
       │  1. POST /api/auth/device/code    │                                   │
       │──────────────────────────────────►│                                   │
       │                                   │                                   │
       │                                   │  存入 Redis (15分钟 TTL)          │
       │                                   │  ═══════════════════════          │
       │                                   │                                   │
       │  返回 device_code, user_code      │                                   │
       │◄──────────────────────────────────│                                   │
       │                                   │                                   │
       │  2. 显示提示:                     │                                   │
       │  "请访问 xxx.com/device"          │                                   │
       │  "输入代码: ABCD-1234"            │                                   │
       │  并自动打开浏览器                  │                                   │
       │  ════════════════════════════════ │                                   │
       │                                   │                                   │
       │                                   │  3. 用户打开页面，登录账号         │
       │                                   │◄─────────────────────────────────│
       │                                   │                                   │
       │                                   │  4. 输入 user_code: ABCD-1234     │
       │                                   │◄─────────────────────────────────│
       │                                   │                                   │
       │                                   │  5. POST /api/auth/device/authorize
       │                                   │◄─────────────────────────────────│
       │                                   │                                   │
       │                                   │  更新 Redis: status = authorized  │
       │                                   │  ═════════════════════════════════│
       │                                   │                                   │
       │  6. GET /api/auth/device/status   │                                   │
       │     (轮询，每5秒一次)              │                                   │
       │──────────────────────────────────►│                                   │
       │                                   │                                   │
       │                                   │  从 Redis 读取状态                │
       │                                   │  ═════════════════                │
       │                                   │                                   │
       │  返回 access_token, desktop_id    │                                   │
       │◄──────────────────────────────────│                                   │
       │                                   │                                   │
       │  7. 保存 token 到本地配置文件      │                                   │
       │  ════════════════════════════════ │                                   │
       │                                   │                                   │
       │  8. 连接 WebSocket                │                                   │
       │──────────────────────────────────►│                                   │
       │                                   │                                   │
       │                                   │  更新 Redis: 设备在线             │
       │                                   │  ═════════════════════            │
       │                                   │                                   │
       │  ✓ 登录成功，开始服务             │                                   │
       ▼                                   ▼                                   ▼
```

### 6.2 消息收发流程

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│    手机端    │         │    服务端    │         │    电脑端    │
└──────┬───────┘         └──────┬───────┘         └──────┬───────┘
       │                        │                        │
       │  1. WS: user:message   │                        │
       │   {desktop_id: 1,      │                        │
       │    content: "写登录页"}│                        │
       │───────────────────────►│                        │
       │                        │                        │
       │                        │  2. 检查 Redis 设备是否在线
       │                        │  3. 保存消息到 MySQL    │
       │                        │                        │
       │                        │  4. WS: user:message   │
       │                        │───────────────────────►│
       │                        │                        │
       │                        │                        │  5. 转发到 AgentAPI
       │                        │                        │     POST localhost:3284/message
       │                        │                        │
       │                        │  6. WS: agent:status   │
       │  7. WS: agent:status   │     {status: "running"}│
       │   {status: "running"}  │◄───────────────────────│
       │◄───────────────────────│                        │
       │                        │                        │
       │                        │                        │  8. 订阅 AgentAPI SSE
       │                        │                        │     监听 AI 输出
       │                        │                        │
       │                        │  9. WS: agent:stream   │
       │  10. WS: agent:stream  │     {delta: "好的，"}  │
       │      显示打字机效果    │◄───────────────────────│
       │◄───────────────────────│                        │
       │                        │                        │
       │                        │  ... 多次流式输出 ...  │
       │                        │                        │
       │                        │  11. WS: agent:response│
       │  12. WS: agent:response│     {完整响应}         │
       │      显示完整结果      │◄───────────────────────│
       │◄───────────────────────│                        │
       │                        │  13. 保存响应到 MySQL  │
       │                        │                        │
       │                        │  14. WS: agent:status  │
       │  15. WS: agent:status  │     {status: "idle"}   │
       │   {status: "idle"}     │◄───────────────────────│
       │◄───────────────────────│                        │
       ▼                        ▼                        ▼
```

### 6.3 心跳与离线检测

```
┌──────────────┐                    ┌──────────────┐
│    电脑端    │                    │    服务端    │
└──────┬───────┘                    └──────┬───────┘
       │                                   │
       │  WS: heartbeat (每30秒)           │
       │──────────────────────────────────►│
       │                                   │
       │                                   │  更新 Redis:
       │                                   │  desktop:{id}:heartbeat = now
       │                                   │  TTL = 2分钟
       │                                   │
       │  ... 30秒后 ...                   │
       │                                   │
       │  WS: heartbeat                    │
       │──────────────────────────────────►│
       │                                   │
       │                                   │  更新 Redis TTL
       │                                   │
       │                                   │
       │  === 如果电脑端断开 ===           │
       │                                   │
       │  (WebSocket 断开事件)             │
       │  ════════════════════════════════►│
       │                                   │
       │                                   │  从 Redis 移除在线状态:
       │                                   │  SREM online:desktops {id}
       │                                   │  DEL desktop:{id}:heartbeat
       │                                   │
       │                                   │  通知用户的手机端:
       │                                   │  WS: desktop:offline
       │                                   │──────────────────────►📱
       │                                   │
       ▼                                   ▼
```

---

## 七、部署方案

### 7.1 配置文件

```yaml
# configs/config.yaml
server:
  port: 8080
  mode: release  # debug / release

mysql:
  host: your-rds.mysql.rds.aliyuncs.com
  port: 3306
  user: your_user
  password: your_password
  database: remote_claude
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: your-redis.redis.rds.aliyuncs.com
  port: 6379
  password: your_password
  db: 0
  pool_size: 100

jwt:
  secret: your-jwt-secret-key-at-least-32-chars
  access_expire: 24h
  refresh_expire: 168h  # 7 days

log:
  level: info
  format: json
```

### 7.2 Docker 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  server:
    build: ./server
    ports:
      - "8080:8080"
    environment:
      - CONFIG_FILE=/app/configs/config.prod.yaml
    volumes:
      - ./configs:/app/configs:ro
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  mobile:
    build: ./mobile
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - server
    restart: always
```

```dockerfile
# server/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/server .
EXPOSE 8080
CMD ["./server"]
```

### 7.3 电脑端分发

```makefile
# desktop/Makefile

VERSION := 1.0.0
BINARY := remote-claude

.PHONY: build-all
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-windows-amd64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o dist/$(BINARY)-darwin-amd64 ./cmd/remote-claude

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o dist/$(BINARY)-darwin-arm64 ./cmd/remote-claude

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o dist/$(BINARY)-linux-amd64 ./cmd/remote-claude

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o dist/$(BINARY)-windows-amd64.exe ./cmd/remote-claude
```

用户安装脚本：
```bash
#!/bin/bash
# install.sh

set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

DOWNLOAD_URL="https://your-domain.com/download/remote-claude-${OS}-${ARCH}"

echo "正在下载 remote-claude..."
curl -fsSL "$DOWNLOAD_URL" -o /usr/local/bin/remote-claude
chmod +x /usr/local/bin/remote-claude

echo "✓ 安装成功！"
echo "运行 'remote-claude login' 开始使用"
```

---

## 八、开发计划

### 8.1 第一阶段：MVP（2-3 周）

| 任务 | 预计时间 | 说明 |
|------|---------|------|
| 服务端基础框架 | 2天 | Gin + GORM + MySQL + Redis 连接 |
| 用户认证模块 | 2天 | 注册、登录、JWT、黑名单 |
| 设备登录流程 | 2天 | Device Flow + Redis 缓存 |
| WebSocket Hub | 3天 | 连接管理、消息路由、Redis 状态同步 |
| 电脑端 CLI | 3天 | 登录、连接、消息转发 |
| AgentAPI 集成 | 2天 | Claude Code 控制 |
| 手机端基础页面 | 3天 | 登录、设备列表、对话 |

**MVP 交付物：**
- 用户可以注册登录
- 电脑端可以扫码授权
- 手机可以发送指令并收到响应
- 实时显示设备在线状态

### 8.2 第二阶段：完善（1-2 周）

| 任务 | 说明 |
|------|------|
| 会话管理 | 多会话、历史记录 |
| 流式输出 | 打字机效果 |
| 断线重连 | 自动重连机制 |
| 错误处理 | 友好的错误提示 |
| UI 优化 | 代码高亮、Markdown 渲染 |

### 8.3 第三阶段：扩展（按需）

| 功能 | 说明 |
|------|------|
| 多 AI 工具支持 | Aider、Goose 等 |
| 云电脑端 | 服务端托管的开发环境 |
| 团队协作 | 多人共享设备 |
| 通知推送 | 任务完成通知 |

---

## 九、关键依赖

### 9.1 服务端依赖 (Go)

```go
// go.mod
module remote-claude-server

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1           // Web 框架
    github.com/gorilla/websocket v1.5.1       // WebSocket
    gorm.io/gorm v1.25.5                      // ORM
    gorm.io/driver/mysql v1.5.2               // MySQL 驱动
    github.com/redis/go-redis/v9 v9.3.0       // Redis 客户端
    github.com/golang-jwt/jwt/v5 v5.2.0       // JWT
    github.com/spf13/viper v1.18.2            // 配置管理
    golang.org/x/crypto v0.17.0               // 密码哈希
    github.com/google/uuid v1.5.0             // UUID 生成
)
```

### 9.2 电脑端依赖 (Go)

```go
// go.mod
module remote-claude-cli

go 1.21

require (
    github.com/spf13/cobra v1.8.0             // CLI 框架
    github.com/spf13/viper v1.18.2            // 配置管理
    github.com/gorilla/websocket v1.5.1       // WebSocket
    github.com/google/uuid v1.5.0             // UUID 生成
)
```

### 9.3 手机端依赖 (React)

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-router-dom": "^6.21.0",
    "zustand": "^4.4.7",
    "axios": "^1.6.2",
    "react-markdown": "^9.0.1",
    "react-syntax-highlighter": "^15.5.0"
  },
  "devDependencies": {
    "vite": "^5.0.8",
    "typescript": "^5.3.3",
    "tailwindcss": "^3.4.0",
    "vite-plugin-pwa": "^0.17.4"
  }
}
```

### 9.4 外部依赖

| 组件 | 用途 | 安装方式 |
|------|------|---------|
| AgentAPI | Claude Code HTTP 封装 | 电脑端自动下载 |
| Claude Code | AI 编程工具 | 用户需预先安装 |

---

## 十、注意事项

### 10.1 安全考虑

| 风险 | 措施 |
|------|------|
| Token 泄露 | JWT 设置合理过期时间，支持刷新，登出加入 Redis 黑名单 |
| 设备冒充 | device_token 唯一且随机，授权码 Redis 15分钟 TTL |
| 消息窃听 | 全程 HTTPS/WSS |
| 越权访问 | 每次请求验证用户和设备归属关系 |
| Redis 数据安全 | 使用阿里云 Redis，开启密码认证 |

### 10.2 性能考虑

| 场景 | 优化措施 |
|------|---------|
| 大量 WebSocket 连接 | Go 协程轻量，单机支撑数万连接 |
| 消息广播 | 使用 channel 异步发送 |
| 数据库压力 | 热点数据使用 Redis 缓存 |
| 在线状态查询 | Redis Set 存储，O(1) 查询 |
| 多实例部署 | Redis Pub/Sub 跨实例广播 |

### 10.3 用户体验

| 场景 | 处理方式 |
|------|---------|
| 电脑离线 | 手机端实时显示离线状态（Redis 状态同步），禁用发送 |
| 网络断开 | 自动重连，显示连接状态 |
| 长时间任务 | 流式输出，实时显示进度 |
| 授权码过期 | Redis TTL 自动清理，提示用户重新获取 |

---

## 附录 A：Go 语言快速入门

### A.1 基础语法速查

```go
// 变量
name := "Claude"                    // 类型推断
var age int = 3                     // 显式类型
var list []string                   // 声明切片

// 函数
func Add(a, b int) int {
    return a + b
}

// 多返回值
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为0")
    }
    return a / b, nil
}

// 调用
result, err := Divide(10, 2)
if err != nil {
    log.Fatal(err)
}

// 结构体（相当于 Java 的 class）
type User struct {
    ID       int64
    Username string
}

// 方法
func (u *User) GetDisplayName() string {
    return u.Username
}

// 接口（隐式实现）
type Repository interface {
    FindByID(id int64) (*User, error)
}
```

### A.2 常用命令

```bash
# 初始化项目
go mod init your-module-name

# 下载依赖
go mod tidy

# 运行
go run main.go

# 编译
go build -o app main.go

# 测试
go test ./...

# 格式化代码
go fmt ./...
```

---

## 附录 B：Redis 命令速查

```bash
# 字符串
SET key value EX 900           # 设置，15分钟过期
GET key                        # 获取
DEL key                        # 删除

# 哈希
HSET user:1 name zhang age 18  # 设置多个字段
HGETALL user:1                 # 获取所有字段
HGET user:1 name               # 获取单个字段

# 集合
SADD online:desktops 1 2 3     # 添加成员
SMEMBERS online:desktops       # 获取所有成员
SISMEMBER online:desktops 1    # 检查成员是否存在
SREM online:desktops 1         # 移除成员

# 过期时间
EXPIRE key 900                 # 设置过期时间（秒）
TTL key                        # 查看剩余时间

# 发布订阅
PUBLISH channel message        # 发布消息
SUBSCRIBE channel              # 订阅频道
```

---

**文档版本**: 2.0  
**更新内容**: 新增 Redis 集成方案  
**最后更新**: 2024年1月
