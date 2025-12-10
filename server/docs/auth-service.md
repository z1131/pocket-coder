# 认证服务 (AuthService)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

认证服务处理用户注册、登录、登出以及设备授权流程。

## 主要功能

### 1. 用户注册

流程：
1. 验证用户名/邮箱是否已存在
2. 对密码进行 bcrypt 哈希
3. 创建用户记录
4. 返回用户信息

### 2. 用户登录

流程：
1. 根据用户名查找用户
2. 验证密码
3. 生成 Access Token 和 Refresh Token
4. 返回 Token 和用户信息

### 3. 用户登出

流程：
1. 将当前 Token 加入 Redis 黑名单
2. 黑名单 TTL 设为 Token 剩余有效期

### 4. Token 刷新

流程：
1. 验证 Refresh Token
2. 生成新的 Access Token
3. 返回新 Token

### 5. 设备授权流程 (Device Flow)

```
电脑端                    服务端                    手机端
   |                        |                        |
   |  1. 请求设备码         |                        |
   |----------------------->|                        |
   |                        |                        |
   |  2. 返回设备码+用户码   |                        |
   |<-----------------------|                        |
   |                        |                        |
   |  显示: "请访问 xxx.com  |                        |
   |         输入: ABCD-1234"|                        |
   |                        |                        |
   |                        |  3. 用户输入用户码      |
   |                        |<-----------------------|
   |                        |                        |
   |                        |  4. 授权设备码          |
   |                        |<-----------------------|
   |                        |                        |
   |  5. 轮询获取授权状态    |                        |
   |----------------------->|                        |
   |                        |                        |
   |  6. 返回 Token + 设备ID|                        |
   |<-----------------------|                        |
```

## 方法列表

| 方法 | 说明 |
|------|------|
| Register(req) | 用户注册 |
| Login(req) | 用户登录 |
| Logout(tokenHash, expireAt) | 用户登出 |
| RefreshToken(refreshToken) | 刷新 Token |
| RequestDeviceCode(req) | 请求设备授权码 |
| GetDeviceStatus(deviceCode) | 获取设备授权状态 |
| AuthorizeDevice(userID, userCode) | 授权设备 |

## 文件路径

- `internal/service/auth_service.go`

## 依赖

- UserRepository: 用户数据访问
- DesktopRepository: 设备数据访问
- RedisCache: Redis 缓存操作
- JWTService: JWT Token 生成和验证
