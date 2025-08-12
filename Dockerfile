# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpc-kit ./cmd/grpc-kit

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN addgroup -g 1001 -S grpckit && \
    adduser -u 1001 -S grpckit -G grpckit

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/grpc-kit .
COPY --from=builder /app/config ./config

# 切换到非 root 用户
USER grpckit

# 暴露端口
EXPOSE 8080 9090 8081

# 运行应用
CMD ["./grpc-kit"]