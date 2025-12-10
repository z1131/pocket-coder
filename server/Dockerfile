# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制 go.mod
COPY go.mod ./

# 复制源代码
COPY . .

# 下载依赖并整理
RUN go mod tidy

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 运行阶段（使用轻量级镜像）
FROM alpine:latest

# 安装 ca-certificates（用于 HTTPS 请求）和 tzdata（时区）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件目录
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]
