# Nginx客户端演示

本示例演示如何在 `go-grpc-kit` 框架中配置gRPC客户端以连接到Nginx地址。

## 配置说明

### 1. 启用DNS解析器

要连接到Nginx地址，需要在配置文件中**省略或清空** `discovery` 配置部分，这样框架会自动使用gRPC内置的DNS解析器：

```yaml
# 注意：没有配置discovery部分，这样会自动使用DNS解析器
grpc:
  client:
    timeout: 30
    load_balancing: "round_robin"
    # ... 其他客户端配置
```

### 2. 支持的Nginx地址格式

- **域名地址**: `nginx.example.com:443`
- **IP地址**: `192.168.1.100:80`
- **本地地址**: `localhost:8080`
- **内部负载均衡器**: `nginx-lb.internal:9090`

### 3. 客户端配置选项

```yaml
grpc:
  client:
    # 连接配置
    timeout: 30                    # 连接超时时间（秒）
    max_recv_msg_size: 4194304     # 最大接收消息大小（4MB）
    max_send_msg_size: 4194304     # 最大发送消息大小（4MB）
    
    # Keepalive配置
    keepalive_time: 30             # Keepalive时间（秒）
    keepalive_timeout: 5           # Keepalive超时（秒）
    permit_without_stream: true    # 允许无流时发送keepalive
    
    # 负载均衡策略
    load_balancing: "round_robin"  # round_robin, pick_first
    
    # 重试策略
    retry_policy:
      max_attempts: 3              # 最大重试次数
      initial_backoff: "1s"        # 初始退避时间
      max_backoff: "30s"           # 最大退避时间
      backoff_multiplier: 2.0      # 退避倍数
      retryable_status_codes:      # 可重试的状态码
        - "UNAVAILABLE"
        - "DEADLINE_EXCEEDED"
        - "RESOURCE_EXHAUSTED"
    
    # 压缩配置
    compression: "gzip"            # 压缩算法：none, gzip
    
    # 拦截器配置
    enable_logging: true           # 启用日志拦截器
    enable_metrics: true           # 启用指标拦截器
```

## 使用方法

### 1. 在代码中连接Nginx

```go
// 创建应用程序
application := app.New(app.WithConfig(cfg))

// 连接到Nginx地址
client, err := application.GetClient("nginx.example.com:443")
if err != nil {
    log.Printf("连接失败: %v", err)
    return
}

// 使用客户端调用gRPC服务
// userClient := pb.NewUserServiceClient(client)
// response, err := userClient.GetUser(ctx, &pb.GetUserRequest{...})
```

### 2. 连接到Nginx上游服务

如果Nginx配置了多个上游服务，可以通过同一个Nginx地址连接到不同的服务：

```go
// 所有服务都通过同一个Nginx网关
nginxGateway := "nginx-gateway.example.com:443"

// 连接到用户服务
userClient, err := application.GetClient(nginxGateway)
if err != nil {
    log.Printf("连接用户服务失败: %v", err)
}

// 连接到订单服务  
orderClient, err := application.GetClient(nginxGateway)
if err != nil {
    log.Printf("连接订单服务失败: %v", err)
}
```

## Nginx配置示例

### 1. HTTP/2 gRPC代理配置

```nginx
upstream grpc_backend {
    server backend1.example.com:9090;
    server backend2.example.com:9090;
    server backend3.example.com:9090;
}

server {
    listen 443 ssl http2;
    server_name nginx.example.com;
    
    # SSL配置
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        grpc_pass grpc://grpc_backend;
        grpc_set_header Host $host;
        grpc_set_header X-Real-IP $remote_addr;
        grpc_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        grpc_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 多服务路由配置

```nginx
upstream user_service {
    server user1.example.com:9090;
    server user2.example.com:9090;
}

upstream order_service {
    server order1.example.com:9090;
    server order2.example.com:9090;
}

server {
    listen 443 ssl http2;
    server_name nginx-gateway.example.com;
    
    # 用户服务路由
    location /user.UserService/ {
        grpc_pass grpc://user_service;
    }
    
    # 订单服务路由
    location /order.OrderService/ {
        grpc_pass grpc://order_service;
    }
}
```

## 运行示例

```bash
# 进入示例目录
cd examples/nginx_client_demo

# 运行示例
go run main.go
```

## 工作原理

1. **DNS解析器启用**: 当配置文件中没有 `discovery` 配置时，框架自动使用gRPC内置DNS解析器
2. **直接连接**: 客户端直接连接到指定的Nginx地址，不经过服务发现
3. **负载均衡**: 如果Nginx后面有多个后端服务，负载均衡由Nginx处理
4. **连接复用**: 框架会复用到同一地址的连接，提高性能

## 注意事项

1. **TLS配置**: 如果Nginx使用HTTPS，确保客户端配置了正确的TLS设置
2. **健康检查**: 如果Nginx支持gRPC健康检查协议，可以使用框架的健康检查功能
3. **超时设置**: 根据网络环境和Nginx配置调整客户端超时时间
4. **重试策略**: 配置合适的重试策略以处理网络波动和Nginx重启等情况