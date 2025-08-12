# gRPC 框架配置指南

本文档详细介绍了 gRPC 框架的配置选项，包括新增的 gRPC 原生配置支持。

## 配置结构

### 服务器配置 (server)

```yaml
server:
  host: "0.0.0.0"        # 服务器监听地址
  port: 8080             # HTTP 端口
  grpc_port: 9090        # gRPC 端口
```

### gRPC 配置 (grpc)

#### 服务器配置 (grpc.server)

##### 消息大小限制
```yaml
grpc:
  server:
    max_recv_msg_size: 4194304  # 最大接收消息大小 (字节)，默认 4MB
    max_send_msg_size: 4194304  # 最大发送消息大小 (字节)，默认 4MB
```

##### 连接配置
```yaml
grpc:
  server:
    max_concurrent_streams: 1000  # 最大并发流数量，默认 100
    connection_timeout: 30        # 连接超时时间 (秒)，默认 30
```

##### Keepalive 配置
```yaml
grpc:
  server:
    keepalive_time: 30      # Keepalive 时间间隔 (秒)，默认 30
    keepalive_timeout: 5    # Keepalive 超时时间 (秒)，默认 5
    keepalive_min_time: 5   # 最小 Keepalive 时间 (秒)，默认 5
```

##### 安全配置
```yaml
grpc:
  server:
    enable_reflection: true  # 是否启用反射服务，默认 true
```

##### 压缩配置
```yaml
grpc:
  server:
    enable_compression: true  # 是否启用压缩，默认 true
    compression_level: 6      # 压缩级别 (1-9)，默认 6
```

##### 拦截器配置
```yaml
grpc:
  server:
    enable_logging: true   # 是否启用日志拦截器，默认 true
    enable_metrics: true   # 是否启用指标拦截器，默认 true
    enable_recovery: true  # 是否启用恢复拦截器，默认 true
    enable_tracing: false  # 是否启用追踪拦截器，默认 false
```

#### 客户端配置 (grpc.client)

##### 消息大小限制
```yaml
grpc:
  client:
    max_recv_msg_size: 4194304  # 最大接收消息大小 (字节)，默认 4MB
    max_send_msg_size: 4194304  # 最大发送消息大小 (字节)，默认 4MB
```

##### 连接配置
```yaml
grpc:
  client:
    connection_timeout: 30  # 连接超时时间 (秒)，默认 30
```

##### Keepalive 配置
```yaml
grpc:
  client:
    keepalive_time: 30           # Keepalive 时间间隔 (秒)，默认 30
    keepalive_timeout: 5         # Keepalive 超时时间 (秒)，默认 5
    permit_without_stream: false # 是否允许无流时发送 Keepalive，默认 false
```

##### 负载均衡配置
```yaml
grpc:
  client:
    load_balancing_policy: "round_robin"  # 负载均衡策略，默认 "round_robin"
```

支持的负载均衡策略：
- `round_robin`: 轮询
- `pick_first`: 选择第一个可用的
- `grpclb`: gRPC 负载均衡器

##### 重试策略配置
```yaml
grpc:
  client:
    retry_policy:
      max_attempts: 3                                        # 最大重试次数，默认 3
      initial_backoff: "1s"                                 # 初始退避时间，默认 1s
      max_backoff: "10s"                                     # 最大退避时间，默认 10s
      backoff_multiplier: 2.0                               # 退避倍数，默认 2.0
      retryable_status_codes: ["UNAVAILABLE", "DEADLINE_EXCEEDED"]  # 可重试的状态码
```

支持的重试状态码：
- `CANCELLED`
- `UNKNOWN`
- `INVALID_ARGUMENT`
- `DEADLINE_EXCEEDED`
- `NOT_FOUND`
- `ALREADY_EXISTS`
- `PERMISSION_DENIED`
- `RESOURCE_EXHAUSTED`
- `FAILED_PRECONDITION`
- `ABORTED`
- `OUT_OF_RANGE`
- `UNIMPLEMENTED`
- `INTERNAL`
- `UNAVAILABLE`
- `DATA_LOSS`
- `UNAUTHENTICATED`

##### 压缩配置
```yaml
grpc:
  client:
    enable_compression: true  # 是否启用压缩，默认 true
    compression_level: 6      # 压缩级别 (1-9)，默认 6
```

##### 拦截器配置
```yaml
grpc:
  client:
    enable_logging: true   # 是否启用日志拦截器，默认 true
    enable_metrics: true   # 是否启用指标拦截器，默认 true
    enable_tracing: false  # 是否启用追踪拦截器，默认 false
```

### 服务发现配置 (discovery)

#### 使用服务发现
```yaml
discovery:
  type: "etcd"           # 服务发现类型，支持 "etcd", "consul"
  endpoints:             # 服务发现端点列表
    - "localhost:2379"
  timeout: 5             # 连接超时时间 (秒)，默认 5
  namespace: "grpc"      # 命名空间，默认为空
```

支持的服务发现类型：
- `etcd`: 使用 etcd 作为服务注册中心
- `consul`: 使用 Consul 作为服务注册中心

#### 使用DNS解析器
当不配置 `discovery` 部分或将 `type` 设置为空字符串时，客户端将自动使用 gRPC 内置的 DNS 解析器：

```yaml
# 方式1: 完全不配置 discovery 部分
# discovery 部分被省略

# 方式2: 显式设置为空
discovery:
  type: ""               # 空字符串表示使用 DNS 解析器
```

使用 DNS 解析器时，客户端连接目标应该是 `host:port` 格式：
- `example.com:9090` - 域名
- `192.168.1.100:9090` - IP地址
- `localhost:9090` - 本地地址

#### DNS解析器 vs 服务发现对比

| 特性 | DNS解析器 | 服务发现 |
|------|-----------|----------|
| 配置复杂度 | 简单 | 复杂 |
| 基础设施依赖 | DNS服务器 | etcd/Consul等 |
| 动态服务发现 | 有限 | 完整支持 |
| 负载均衡 | DNS轮询 | 多种策略 |
| 健康检查 | 无 | 支持 |
| 适用场景 | 简单部署、开发测试 | 微服务架构、生产环境 |

### 日志配置 (logging)

```yaml
logging:
  level: "info"          # 日志级别：debug, info, warn, error
  format: "json"         # 日志格式：json, text
  output: "stdout"       # 日志输出：stdout, stderr, 文件路径
```

### TLS 配置 (tls)

```yaml
tls:
  enabled: false         # 是否启用 TLS
  cert_file: ""          # 证书文件路径
  key_file: ""           # 私钥文件路径
  ca_file: ""            # CA 证书文件路径
```

### 指标配置 (metrics)

```yaml
metrics:
  enabled: true          # 是否启用指标收集
  port: 8081            # 指标服务端口
  path: "/metrics"      # 指标端点路径
```

### 自动注册配置 (auto_register)

```yaml
auto_register:
  enabled: true          # 是否启用自动注册
  scan_dirs:             # 扫描目录列表
    - "./internal/service"
  patterns:              # 文件匹配模式
    - "*.go"
  excludes:              # 排除模式
    - "*_test.go"
  service_name: "example-service"  # 服务名称
```

## 配置优先级

配置的加载优先级（从高到低）：
1. 命令行参数
2. 环境变量
3. 配置文件
4. 默认值

## 环境变量

所有配置项都可以通过环境变量设置，格式为 `GRPC_KIT_` + 配置路径（用下划线分隔，全大写）。

例如：
- `GRPC_KIT_SERVER_HOST=0.0.0.0`
- `GRPC_KIT_GRPC_SERVER_MAX_RECV_MSG_SIZE=8388608`
- `GRPC_KIT_GRPC_CLIENT_RETRY_POLICY_MAX_ATTEMPTS=5`

## 配置验证

框架会在启动时验证配置的有效性：
- 端口号范围检查
- 超时时间合理性检查
- 文件路径存在性检查
- 枚举值有效性检查

## 最佳实践

1. **生产环境建议**：
   - 启用 TLS
   - 设置合理的超时时间
   - 启用指标收集
   - 禁用反射服务

2. **开发环境建议**：
   - 启用反射服务
   - 启用详细日志
   - 设置较短的超时时间

3. **性能优化**：
   - 根据业务需求调整消息大小限制
   - 合理设置并发流数量
   - 启用压缩以减少网络传输

4. **监控和调试**：
   - 启用指标收集
   - 配置适当的日志级别
   - 使用追踪拦截器进行问题排查

## 使用示例

### 使用DNS解析器的完整配置示例

```yaml
# 使用DNS解析器的配置示例
server:
  host: "0.0.0.0"
  grpcPort: 9090

grpc:
  server:
    maxRecvMsgSize: 4194304
    maxSendMsgSize: 4194304
    maxConcurrentStreams: 1000
    keepalive:
      maxConnectionIdle: 300s
      maxConnectionAge: 600s
      maxConnectionAgeGrace: 30s
      time: 60s
      timeout: 5s
    enableReflection: true
    enableRecovery: true
    
  client:
    timeout: 30
    loadBalancing: "round_robin"
    maxRecvMsgSize: 4194304
    maxSendMsgSize: 4194304
    keepaliveTime: 30
    keepaliveTimeout: 5
    permitWithoutStream: true
    enableLogging: true
    retryPolicy:
      maxAttempts: 3
      initialBackoff: "1s"
      maxBackoff: "30s"
      backoffMultiplier: 2.0
      retryableStatusCodes: ["UNAVAILABLE", "DEADLINE_EXCEEDED"]

logging:
  level: "info"
  format: "json"

metrics:
  enabled: false
  port: 8081
  path: "/metrics"

# 注意：没有配置discovery部分，客户端将使用DNS解析器
```

### 代码使用示例

```go
package main

import (
    "github.com/go-grpc-kit/go-grpc-kit/pkg/app"
    "github.com/go-grpc-kit/go-grpc-kit/pkg/config"
)

func main() {
    // 加载配置
    cfg, err := config.Load("./config/application.yml")
    if err != nil {
        log.Fatal(err)
    }

    // 创建应用程序
    application := app.New(app.WithConfig(cfg))

    // 获取客户端连接（使用DNS解析）
    conn, err := application.GetClient("example.com:9090")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // 使用连接创建gRPC客户端
    client := your_proto.NewYourServiceClient(conn)
    
    // 调用服务方法
    resp, err := client.YourMethod(context.Background(), &your_proto.Request{})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %v", resp)
}
```