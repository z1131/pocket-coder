# 配置模块 (Config)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

配置模块负责加载和管理应用程序的所有配置项，包括：
- 服务器配置（端口、运行模式）
- MySQL 数据库配置
- Redis 配置
- JWT 配置
- 日志配置

## 技术方案

使用 `spf13/viper` 库进行配置管理，支持：
- YAML 配置文件
- 环境变量覆盖
- 配置热更新（可选）

## 配置项说明

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| server.port | int | 服务端口 | 8080 |
| server.mode | string | 运行模式(debug/release) | debug |
| mysql.host | string | MySQL 主机地址 | localhost |
| mysql.port | int | MySQL 端口 | 3306 |
| mysql.user | string | MySQL 用户名 | root |
| mysql.password | string | MySQL 密码 | - |
| mysql.database | string | 数据库名 | pocket_coder |
| redis.host | string | Redis 主机地址 | localhost |
| redis.port | int | Redis 端口 | 6379 |
| redis.password | string | Redis 密码 | - |
| redis.db | int | Redis 数据库索引 | 0 |
| jwt.secret | string | JWT 密钥 | - |
| jwt.access_expire | duration | Access Token 过期时间 | 24h |
| jwt.refresh_expire | duration | Refresh Token 过期时间 | 168h |

## 文件路径

- 配置定义: `internal/config/config.go`
- 配置模板: `configs/config.yaml`
- 生产配置: `configs/config.prod.yaml`

## 使用示例

```go
// 加载配置
cfg, err := config.Load("configs/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 使用配置
fmt.Println(cfg.Server.Port)
fmt.Println(cfg.MySQL.Host)
```

## 注意事项

1. 生产环境密码等敏感信息建议使用环境变量
2. JWT Secret 至少 32 个字符
3. 确保 MySQL 和 Redis 连接信息正确
