# 中间件模块 (Middleware)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

中间件层提供 HTTP 请求的预处理和后处理功能，包括：
- JWT 认证
- CORS 跨域处理
- 请求日志记录

## 中间件列表

### AuthMiddleware (JWT 认证)

功能：
- 从请求头提取 Bearer Token
- 验证 Token 有效性
- 检查 Token 是否在黑名单
- 将用户信息注入请求上下文

流程：
```
Request → Extract Token → Validate JWT → Check Blacklist → Set User Context → Next Handler
```

使用方式：
```go
// 需要认证的路由组
auth := r.Group("/api")
auth.Use(middleware.AuthMiddleware(jwtService, cache))
```

### CORSMiddleware (跨域处理)

功能：
- 允许指定来源的跨域请求
- 支持预检请求 (OPTIONS)
- 配置允许的请求头和方法

配置项：
- AllowOrigins: 允许的来源
- AllowMethods: 允许的 HTTP 方法
- AllowHeaders: 允许的请求头
- AllowCredentials: 是否允许携带凭据

### LoggerMiddleware (请求日志)

功能：
- 记录请求方法、路径、状态码
- 记录请求耗时
- 记录客户端 IP

日志格式：
```
[GIN] 2024/01/15 - 10:30:00 | 200 |    12.345ms | 192.168.1.1 | GET /api/user/profile
```

## 上下文中的用户信息

认证中间件会将以下信息存入 Gin 上下文：

```go
// 获取用户 ID
userID := c.GetInt64("user_id")

// 获取用户名
username := c.GetString("username")
```

## 文件路径

- `internal/middleware/auth.go`
- `internal/middleware/cors.go`
- `internal/middleware/logger.go`

## 注意事项

1. 认证中间件依赖 JWT 服务和 Redis 缓存
2. CORS 配置需要根据实际部署环境调整
3. 生产环境建议关闭详细日志以提高性能
