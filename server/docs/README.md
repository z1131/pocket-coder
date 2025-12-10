# Pocket Coder 服务端开发文档

## 项目概述

这是 Pocket Coder 的服务端实现，采用 Go + Gin + GORM + Redis 技术栈。服务端作为手机端和电脑端的桥梁，负责：

1. 用户认证与授权
2. 设备管理
3. 会话管理
4. WebSocket 消息路由

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.21+ |
| Web框架 | Gin | 1.9.1 |
| ORM | GORM | 1.25.5 |
| 数据库 | MySQL | 8.0 |
| 缓存 | Redis | 7.0 |
| WebSocket | gorilla/websocket | 1.5.1 |
| JWT | golang-jwt | v5 |
| 配置 | viper | 1.18.2 |

## 项目结构

```
server/
├── cmd/
│   └── server/
│       └── main.go                 # 入口文件
├── internal/
│   ├── config/                     # 配置管理
│   ├── model/                      # 数据模型
│   ├── repository/                 # 数据访问层
│   ├── cache/                      # Redis 缓存层
│   ├── service/                    # 业务逻辑层
│   ├── handler/                    # HTTP 处理器
│   ├── websocket/                  # WebSocket 模块
│   └── middleware/                 # 中间件
├── pkg/                            # 公共工具包
├── configs/                        # 配置文件
├── scripts/                        # 脚本文件
└── docs/                           # 开发文档
```

## 模块开发状态

| 模块 | 状态 | 文档 |
|------|------|------|
| 配置模块 | ✅ 已完成 | [config.md](./config.md) |
| 数据模型 | ✅ 已完成 | [model.md](./model.md) |
| 数据访问层 | ✅ 已完成 | [repository.md](./repository.md) |
| Redis缓存 | ✅ 已完成 | [cache.md](./cache.md) |
| 中间件 | ✅ 已完成 | [middleware.md](./middleware.md) |
| 认证服务 | ✅ 已完成 | [auth-service.md](./auth-service.md) |
| 用户服务 | ✅ 已完成 | [user-service.md](./user-service.md) |
| 设备服务 | ✅ 已完成 | [desktop-service.md](./desktop-service.md) |
| 会话服务 | ✅ 已完成 | [session-service.md](./session-service.md) |
| HTTP Handler | ✅ 已完成 | [handler.md](./handler.md) |
| WebSocket Hub | ✅ 已完成 | [websocket.md](./websocket.md) |

## 快速开始

### 1. 安装依赖

```bash
cd server
go mod tidy
```

### 2. 配置数据库

修改 `configs/config.yaml` 中的数据库连接信息。

### 3. 运行服务

```bash
go run cmd/server/main.go
```

## API 文档

详见 [api.md](./api.md)

## 更新日志

- 2024-01-15: 项目初始化
