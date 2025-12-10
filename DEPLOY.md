# Pocket Coder 部署指南

## Docker 部署（推荐）

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+

### 快速启动

```bash
# 1. 进入项目根目录
cd pocket-coder

# 2. 启动所有服务（首次启动会构建镜像）
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f server
```

### 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| server | 8080 | Go 后端服务 |
| mysql | 3306 | MySQL 数据库 |
| redis | 6379 | Redis 缓存 |

### 环境变量配置

在 `docker-compose.yml` 中修改环境变量：

```yaml
environment:
  # 重要：生产环境必须修改
  - JWT_SECRET=your-super-secret-jwt-key-change-in-production
  - MYSQL_PASSWORD=your-secure-password
```

### 常用命令

```bash
# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build

# 清除数据重新开始
docker-compose down -v
docker-compose up -d

# 进入 MySQL 命令行
docker exec -it pocket-coder-mysql mysql -u pocket_coder -p

# 进入 Redis 命令行
docker exec -it pocket-coder-redis redis-cli

# 查看服务器日志
docker logs -f pocket-coder-server
```

### API 测试

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "123456", "nickname": "测试用户"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "123456"}'
```

### 生产部署注意事项

1. **修改密钥**：更改 `JWT_SECRET` 和数据库密码
2. **HTTPS**：使用 Nginx 或 Traefik 配置 SSL
3. **备份**：定期备份 MySQL 数据
4. **监控**：添加 Prometheus + Grafana 监控
5. **日志**：配置日志收集（ELK 或 Loki）

## 本地开发

如果需要本地开发，安装 Go 1.21+：

```bash
# macOS
brew install go

# 安装依赖
cd server
go mod tidy

# 启动 MySQL 和 Redis（使用 Docker）
docker-compose up -d mysql redis

# 运行服务
go run cmd/server/main.go
```
