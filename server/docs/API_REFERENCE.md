# CloudTerm API 参考文档

本文档基于 `cloudterm-ai` 前端项目的需求以及后端 `server` 的架构编写。定义了前后端交互的 RESTful API 接口规范及 WebSocket 通信协议。

## 基础信息

- **Base URL**: `/api/v1`
- **数据格式**: JSON
- **字符编码**: UTF-8

---

## 1. 认证模块 (Auth)

### 1.1 用户注册
**Endpoint**: `POST /auth/register`

**请求参数**:
```json
{
  "username": "zhangsan",
  "email": "zhangsan@example.com",
  "password": "securepassword123",
  "phone": "13800138000" // 可选，前端目前支持 phone/email
}
```

**响应**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...", // JWT Token
    "user": {
      "id": "u_123456",
      "username": "zhangsan",
      "email": "zhangsan@example.com"
    }
  }
}
```

### 1.2 用户登录
**Endpoint**: `POST /auth/login`

**请求参数**:
```json
{
  "identifier": "zhangsan@example.com", // 支持 用户名/邮箱/手机号
  "password": "securepassword123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "u_123456",
      "username": "zhangsan",
      "email": "zhangsan@example.com"
    }
  }
}
```

### 1.3 刷新 Token
**Endpoint**: `POST /auth/refresh`

**请求参数**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

### 1.4 登出
**Endpoint**: `POST /auth/logout`

**请求头**:
- `Authorization`: `Bearer <token>`

**响应**: `200 OK`

---

## 2. 用户管理 (Users)

### 2.1 获取个人信息
**Endpoint**: `GET /users/me`

**请求头**:
- `Authorization`: `Bearer <token>`

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 1,
    "username": "zhangsan",
    "email": "zhangsan@example.com",
    "avatar": "https://..."
  }
}
```

---

## 3. 设备/桌面管理 (Desktops)
*对应前端 `DeviceListView`*

### 3.1 获取设备列表
**Endpoint**: `GET /desktops`

**请求头**:
- `Authorization`: `Bearer <token>`

**响应**:
```json
{
  "code": 200,
  "data": [
    {
      "id": "d_1",
      "name": "MacBook Pro M2",
      "os": "macos", // 对应后端的 os_info
      "status": "online", // 枚举: online, offline
      "ip": "192.168.1.42", // 可能为空
      "last_active": "2023-10-27T10:00:00Z"
    }
  ]
}
```

### 3.2 连接/添加新设备 (前端按钮触发)
**Endpoint**: `POST /desktops`
*(待设计 - 暂未实现)*
*计划用于生成一次性连接命令或二维码，目前业务逻辑主要依赖 CLI 反向注册*

### 3.3 CLI 设备注册 (反向连接)
*CLI 客户端使用的注册接口*
**Endpoint**: `POST /desktops/register`

**请求头**:
- `Authorization`: `Bearer <token>`

**请求参数**:
```json
{
  "name": "Home Server",
  "device_uuid": "550e8400-e29b-41d4-a716-446655440000", // 客户端持久化的 UUID
  "ip": "192.168.1.200", // 可选
  "os_info": "linux/amd64" // 对应前端的 os
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "desktop_id": 123,
    "desktop_token": "eyJhbGciOiJIUzI1NiIs...", // 桌面专用 Token
    "name": "Home Server"
  }
}
```

### 3.4 获取设备详情
**Endpoint**: `GET /desktops/:id`

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 123,
    "name": "Home Server",
    "status": "online",
    // ...
  }
}
```

### 3.5 更新设备信息
**Endpoint**: `PUT /desktops/:id`

**请求参数**:
```json
{
  "name": "My New Server Name"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    // 更新后的设备对象
  }
}
```

### 3.6 删除设备
**Endpoint**: `DELETE /desktops/:id`

**响应**: `204 No Content`

---

## 4. 会话管理 (Sessions)
*对应前端 `TerminalView` 的会话管理*

### 4.1 创建会话 (启动终端)
**Endpoint**: `POST /sessions`

此接口执行两个操作：
1. 在数据库中创建会话记录。
2. 通过 WebSocket 向目标设备发送指令，要求其启动一个新的 Shell 进程 (PTY)。

**请求头**:
- `Authorization`: `Bearer <token>`

**请求参数**:
```json
{
  "desktop_id": 1, // 必填，目标设备 ID
  "working_dir": "/home/user/project" // 可选，指定启动目录
  // is_default 字段由服务端强制控制，手机端无法通过 API 创建默认会话
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "id": 101,
    "desktop_id": 1,
    "agent_type": "claude-code",
    "is_default": false,
    "status": "active",
    "started_at": "2023-10-27T10:05:00Z"
  }
}
```

### 4.2 获取会话列表
**Endpoint**: `GET /sessions`

获取指定设备的所有会话记录，支持分页。响应中包含最近输出的预览（Base64）。

**请求参数 (Query)**:
- `desktop_id`: (必填) 设备 ID
- `page`: (可选) 页码，默认 1
- `page_size`: (可选) 每页数量，默认 20

**响应**:
```json
{
  "code": 200,
  "data": {
    "sessions": [
      {
        "id": 101,
        "desktop_id": 1,
        "status": "active",
        "preview": "bG9nbi4uLi4=", // Base64 编码的最近 1KB 输出
        "is_default": false,
        "started_at": "..."
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 4.3 获取会话详情
**Endpoint**: `GET /sessions/:id`

**响应**:
```json
{
  "code": 200,
  "data": {
    "session": {
      "id": 101,
      "is_default": false,
      "preview": "...",
      // ... 完整会话字段
    }
  }
}
```

### 4.4 删除会话
**Endpoint**: `DELETE /sessions/:id`

结束并软删除会话。
1. 将数据库状态标记为 `ended`。
2. 发送指令给设备，关闭对应的终端进程。
3. 将 Redis 中的实时日志归档到数据库。

**响应**: `204 No Content`

### 4.5 获取活跃会话
**Endpoint**: `GET /desktops/:id/sessions/active`

获取指定设备当前处于活跃状态的会话（即用户最后交互的会话）。

**响应**:
```json
{
  "code": 200,
  "data": {
    "session": {
      "id": 101,
      // ... 若无活跃会话，session 为 null
    }
  }
}
```

---

## 5. Web Terminal WebSocket 协议
*核心功能，对应前端 `TerminalView` 和 `mockTerminalService` 的实际实现*

**Endpoint**: `WS /ws/terminal?session_id=sess_abc123`
或者
**Endpoint**: `WS /ws/terminal?desktop_id=d_1` (如果简化流程)

### 4.1 消息结构
所有 WebSocket 消息采用 JSON 格式。

### 4.2 客户端发送 (Client -> Server)

**发送命令 (Input)**
```json
{
  "type": "input",
  "content": "ls -la\n" // 包含换行符
}
```

**调整终端大小 (Resize)**
```json
{
  "type": "resize",
  "cols": 80,
  "rows": 24
}
```

**心跳 (Ping)**
```json
{
  "type": "ping"
}
```

### 4.3 服务端发送 (Server -> Client)

**终端输出 (Output)**
```json
{
  "type": "output",
  "content": "total 12\ndrwxr-xr-x 2 user...",
  "timestamp": 1698300000000
}
```

**系统消息 (System/Error)**
*对应前端 `LineType` 中的 `system`, `error`, `info`*
```json
{
  "type": "system", // 或 'error', 'info'
  "content": "Connection established.",
  "timestamp": 1698300000000
}
```

**Pong**
```json
{
  "type": "pong"
}
```

---



```