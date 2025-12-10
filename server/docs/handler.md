# HTTP Handler 模块

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

HTTP Handler 处理 REST API 请求，对应 Java 中的 Controller 层。

## Handler 列表

### AuthHandler (认证处理器)

| 路由 | 方法 | 说明 |
|------|------|------|
| /api/auth/register | POST | 用户注册 |
| /api/auth/login | POST | 用户登录 |
| /api/auth/logout | POST | 用户登出 |
| /api/auth/refresh | POST | 刷新 Token |
| /api/auth/device/code | POST | 获取设备授权码 |
| /api/auth/device/status | GET | 获取设备授权状态 |
| /api/auth/device/authorize | POST | 授权设备 |

### UserHandler (用户处理器)

| 路由 | 方法 | 说明 |
|------|------|------|
| /api/user/profile | GET | 获取用户信息 |
| /api/user/profile | PUT | 更新用户信息 |

### DesktopHandler (设备处理器)

| 路由 | 方法 | 说明 |
|------|------|------|
| /api/desktops | GET | 获取设备列表 |
| /api/desktops/:id | GET | 获取设备详情 |
| /api/desktops/:id | PUT | 更新设备信息 |
| /api/desktops/:id | DELETE | 删除设备 |

### SessionHandler (会话处理器)

| 路由 | 方法 | 说明 |
|------|------|------|
| /api/desktops/:id/sessions | GET | 获取设备的会话列表 |
| /api/desktops/:id/sessions | POST | 创建新会话 |
| /api/sessions/:id | GET | 获取会话详情 |
| /api/sessions/:id | DELETE | 删除会话 |

## 文件路径

- `internal/handler/auth_handler.go`
- `internal/handler/user_handler.go`
- `internal/handler/desktop_handler.go`
- `internal/handler/session_handler.go`

## 请求参数验证

使用 Gin 的 binding 标签进行参数验证：

```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6"`
    Email    string `json:"email" binding:"omitempty,email"`
}
```

## 统一响应格式

```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

## 错误处理

- 400: 参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器内部错误
