# gRPC 配置演示

这个示例演示了 go-grpc-kit 框架的增强 gRPC 配置功能。

## 功能特性

### 服务端配置
- **消息大小限制**: 可配置最大接收和发送消息大小
- **连接管理**: 支持最大并发流数、连接超时等配置
- **Keepalive 参数**: 可配置 keepalive 时间、超时和最小时间
- **反射服务**: 可选择启用 gRPC 反射服务
- **压缩支持**: 支持 gzip 等压缩算法
- **拦截器**: 可选择启用日志记录、指标收集、恢复和追踪拦截器

### 客户端配置
- **连接超时**: 可配置客户端连接超时时间
- **负载均衡**: 支持多种负载均衡策略（round_robin、pick_first等）
- **消息大小限制**: 可配置最大接收和发送消息大小
- **Keepalive 参数**: 客户端 keepalive 配置
- **重试策略**: 完整的重试策略配置，包括最大尝试次数、退避算法等
- **压缩支持**: 客户端压缩配置
- **拦截器**: 客户端拦截器配置

## 运行示例

1. 编译程序：
```bash
cd examples/grpc_config_demo
go build -o grpc_config_demo main.go
```

2. 运行程序：
```bash
./grpc_config_demo
```

程序将显示当前的 gRPC 配置信息，然后启动一个 gRPC 服务器。

## 配置文件

配置文件位于 `config/application.yml`，包含了所有可用的 gRPC 配置选项：

```yaml
grpc:
  server:
    # 消息大小限制
    max_recv_msg_size: 8388608  # 8MB
    max_send_msg_size: 8388608  # 8MB
    
    # 连接配置
    max_concurrent_streams: 200
    connection_timeout: 180     # 3分钟
    keepalive_time: 60         # 1分钟
    keepalive_timeout: 10      # 10秒
    keepalive_min_time: 10     # 10秒
    
    # 功能开关
    enable_reflection: true
    enable_compression: true
    compression_level: "gzip"
    
    # 拦截器配置
    enable_logging: true
    enable_metrics: true
    enable_recovery: true
    enable_tracing: false

  client:
    # 基础配置
    timeout: 60                # 1分钟
    load_balancing: "round_robin"
    
    # 重试策略
    retry_policy:
      max_attempts: 5
      initial_backoff: "1s"
      max_backoff: "60s"
      backoff_multiplier: 2.0
      retryable_status_codes:
        - "UNAVAILABLE"
        - "DEADLINE_EXCEEDED"
        - "RESOURCE_EXHAUSTED"
```

## 输出示例

运行程序后，你将看到类似以下的输出：

```
INFO    gRPC Server Configuration    {"host": "0.0.0.0", "port": 8080, "grpcPort": 9090, "maxRecvMsgSize": 8388608, "maxSendMsgSize": 8388608, "maxConcurrentStreams": 200, "connectionTimeout": "3m0s", "keepaliveTime": "1m0s", "keepaliveTimeout": "10s", "keepaliveMinTime": "10s", "enableReflection": true, "enableCompression": true, "compressionLevel": "gzip", "enableLogging": true, "enableMetrics": true, "enableRecovery": true, "enableTracing": false}

INFO    gRPC Client Configuration    {"timeout": "1m0s", "maxRecvMsgSize": 8388608, "maxSendMsgSize": 8388608, "loadBalancing": "round_robin", "keepaliveTime": "30s", "keepaliveTimeout": "5s", "permitWithoutStream": true, "enableCompression": true, "compressionLevel": "gzip", "enableLogging": true, "enableMetrics": true, "enableTracing": false}

INFO    Retry Policy Configuration   {"maxAttempts": 5, "initialBackoff": "1s", "maxBackoff": "60s", "backoffMultiplier": 2, "retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED", "RESOURCE_EXHAUSTED"]}
```

## 测试配置

你可以修改 `config/application.yml` 文件中的配置值，然后重新运行程序来查看配置的变化。

## 相关文档

- [配置文档](../../docs/configuration.md) - 完整的配置选项说明
- [示例配置](../config/application.yml) - 更多配置示例