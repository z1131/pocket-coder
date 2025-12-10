# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod
COPY go.mod ./

# 下载依赖
RUN go mod download || true

# 复制源代码
COPY . .

# 整理依赖
RUN go mod tidy

# 编译（静态链接）
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o pocket-coder-darwin-arm64 .
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pocket-coder-darwin-amd64 .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pocket-coder-linux-amd64 .
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pocket-coder-windows-amd64.exe .

# 输出阶段
FROM alpine:latest

WORKDIR /output

COPY --from=builder /app/pocket-coder-* ./

CMD ["ls", "-la"]
