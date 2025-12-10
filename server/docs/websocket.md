# WebSocket 模块

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

WebSocket 模块负责处理实时双向通信，包括：
- 连接管理（电脑端和手机端）
- 消息路由
- 状态同步
- 心跳检测

## 架构设计

```
                    ┌─────────────────────────────┐
                    │         WebSocket Hub       │
                    │                             │
                    │  ┌───────────────────────┐  │
                    │  │   Mobile Clients      │  │
        手机端 ──────┼──┤   map[userID]*Client  │  │
                    │  └───────────────────────┘  │
                    │                             │
                    │  ┌───────────────────────┐  │
                    │  │   Desktop Clients     │  │
       电脑端 ──────┼──┤   map[desktopID]*Client│  │
                    │  └───────────────────────┘  │
                    │                             │
                    │  ┌───────────────────────┐  │
                    │  │   Message Router      │  │
                    │  │   - 用户消息 → 设备     │  │
                    │  │   - 设备响应 → 用户     │  │
                    │  └───────────────────────┘  │
                    │                             │
                    └─────────────────────────────┘
```

## 消息类型

### 电脑端 → 服务端

| 类型 | 说明 |
|------|------|
| heartbeat | 心跳 |
| agent:response | AI 完整响应 |
| agent:stream | AI 流式输出 |
| agent:status | AI 状态变更 |

### 服务端 → 电脑端

| 类型 | 说明 |
|------|------|
| user:message | 用户发送的消息 |
| session:create | 创建新会话 |

### 服务端 → 手机端

| 类型 | 说明 |
|------|------|
| desktop:online | 电脑上线通知 |
| desktop:offline | 电脑下线通知 |
| agent:response | AI 响应（转发） |
| agent:stream | AI 流式输出（转发） |
| agent:status | AI 状态（转发） |

### 手机端 → 服务端

| 类型 | 说明 |
|------|------|
| user:message | 发送消息给指定电脑 |

## 消息格式

```typescript
interface WSMessage {
    type: string;        // 消息类型
    payload: object;     // 消息内容
    timestamp: number;   // 时间戳（毫秒）
    message_id?: string; // 消息ID（用于追踪）
}
```

## 文件结构

- `internal/websocket/hub.go` - 连接管理中心
- `internal/websocket/client.go` - 客户端连接
- `internal/websocket/message.go` - 消息定义
- `internal/websocket/handler.go` - 消息处理

## 连接流程

### 电脑端连接

1. 携带 Desktop Token 连接 `/ws/desktop`
2. 验证 Token，提取设备信息
3. 将客户端注册到 Hub
4. 更新 Redis 在线状态
5. 通知用户的手机端设备上线

### 手机端连接

1. 携带 User Token 连接 `/ws/mobile`
2. 验证 Token，提取用户信息
3. 将客户端注册到 Hub
4. 推送用户的设备在线状态

## 心跳机制

- 电脑端每 30 秒发送心跳
- 服务端更新 Redis 心跳时间（TTL 2分钟）
- 如果 2 分钟内没有心跳，自动断开连接
