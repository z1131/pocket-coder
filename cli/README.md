# Pocket Coder CLI

电脑端命令行客户端，用于连接 Pocket Coder 服务器，将手机端的指令转发给本地 AI 编程工具。

## 功能

- 设备码登录（Device Flow）
- WebSocket 长连接
- 接收手机端消息
- 转发给 Claude Code / Cursor 等工具
- 流式返回 AI 响应

## 安装

```bash
cd cli
go build -o pocket-coder ./cmd/pocket-coder
```

## 使用

### 登录
```bash
./pocket-coder login
```
会显示一个 6 位数字码，在手机 App 中输入即可完成绑定。

### 启动服务
```bash
./pocket-coder start
```
连接到服务器，等待手机端指令。

### 查看状态
```bash
./pocket-coder status
```

### 登出
```bash
./pocket-coder logout
```

## 配置

配置文件位于 `~/.pocket-coder/config.yaml`：

```yaml
server:
  url: http://localhost:8080
  ws_url: ws://localhost:8080

device:
  token: ""  # 设备 Token（登录后自动保存）
  name: ""   # 设备名称
```
