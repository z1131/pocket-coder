# Redis 缓存层 (Cache)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

Redis 缓存层负责处理需要快速访问和临时存储的数据，包括：
- 设备授权码管理
- 在线状态管理
- 会话缓存
- JWT 黑名单
- Pub/Sub 消息广播

## Redis Key 设计

### 设备授权码

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `device_code:{code}` | String(JSON) | 15min | 设备码信息 |
| `user_code:{user_code}` | String | 15min | 用户码 -> 设备码映射 |

### 在线状态

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `online:desktops` | Set | - | 所有在线设备ID |
| `desktop:{id}:heartbeat` | String | 2min | 设备心跳时间戳 |
| `user:{id}:online_desktops` | Set | - | 用户的在线设备 |

### 会话缓存

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `session:{id}:cache` | Hash | 1h | 会话信息缓存 |
| `desktop:{id}:active_session` | String | - | 设备的当前活跃会话 |

### JWT 黑名单

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `jwt:blacklist:{token_hash}` | String | Token剩余时间 | 已登出的 Token |

## 主要方法

### 设备授权码

```go
// 创建设备授权码
CreateDeviceCode(ctx, code, info) error

// 获取设备授权码信息
GetDeviceCode(ctx, code) (*DeviceCodeInfo, error)

// 通过用户码获取设备码
GetDeviceCodeByUserCode(ctx, userCode) (string, error)

// 授权设备码
AuthorizeDeviceCode(ctx, code, userID) error
```

### 在线状态

```go
// 设置设备在线
SetDesktopOnline(ctx, desktopID, userID) error

// 设置设备离线
SetDesktopOffline(ctx, desktopID, userID) error

// 更新心跳
UpdateHeartbeat(ctx, desktopID) error

// 检查设备是否在线
IsDesktopOnline(ctx, desktopID) bool

// 获取用户的在线设备
GetUserOnlineDesktops(ctx, userID) ([]int64, error)
```

### JWT 黑名单

```go
// 将 Token 加入黑名单
BlacklistToken(ctx, tokenHash, expireAt) error

// 检查 Token 是否在黑名单
IsTokenBlacklisted(ctx, tokenHash) bool
```

## 文件路径

- `internal/cache/redis.go`

## 注意事项

1. 所有 Redis 操作都需要传入 context，支持超时控制
2. 使用 Pipeline 批量操作提高性能
3. 合理设置 TTL 避免内存泄漏
4. 在线状态以 Redis 为准，MySQL 中的状态仅作参考
