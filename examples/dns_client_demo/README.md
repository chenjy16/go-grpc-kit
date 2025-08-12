# gRPC DNS客户端演示

本示例演示如何使用 `go-grpc-kit` 框架的内置DNS解析器功能来连接gRPC服务。

## 功能特性

- **DNS解析器支持**: 当没有配置服务发现时，自动使用gRPC内置的DNS解析器
- **灵活的目标地址**: 支持域名、IP地址和端口的组合
- **配置驱动**: 通过配置文件控制客户端行为
- **连接管理**: 自动管理连接生命周期

## 运行示例

1. 编译程序：
```bash
go build -o dns_client_demo ./examples/dns_client_demo
```

2. 运行演示：
```bash
./dns_client_demo
```

## 配置说明

在 `config/application.yml` 中，注意以下关键配置：

### 不配置服务发现
```yaml
# 注意：没有配置discovery部分，客户端将使用gRPC内置的DNS解析器
# discovery:
#   type: ""  # 空字符串表示不使用服务发现
```

### 客户端配置
```yaml
grpc:
  client:
    timeout: 30                    # 连接超时（秒）
    loadBalancing: "round_robin"   # 负载均衡策略
    maxRecvMsgSize: 4194304       # 最大接收消息大小
    maxSendMsgSize: 4194304       # 最大发送消息大小
    keepaliveTime: 30             # Keepalive时间
    keepaliveTimeout: 5           # Keepalive超时
    permitWithoutStream: true     # 允许无流的Keepalive
    enableLogging: true           # 启用日志
    retryPolicy:                  # 重试策略
      maxAttempts: 3
      initialBackoff: "1s"
      maxBackoff: "30s"
      backoffMultiplier: 2.0
      retryableStatusCodes: ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
```

## 使用方式

### 1. 通过应用程序获取客户端

```go
// 创建应用程序
app := app.New(app.WithConfig(cfg))

// 获取客户端连接（使用DNS解析）
conn, err := app.GetClient("example.com:9090")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 使用连接创建gRPC客户端
client := your_proto.NewYourServiceClient(conn)
```

### 2. 支持的目标地址格式

- **域名**: `example.com:9090`
- **IP地址**: `192.168.1.100:9090`
- **本地地址**: `localhost:9090` 或 `127.0.0.1:9090`
- **IPv6地址**: `[::1]:9090`

### 3. DNS解析器的优势

- **简单性**: 无需额外的服务发现组件
- **标准化**: 使用标准的DNS协议
- **兼容性**: 与现有的DNS基础设施兼容
- **性能**: DNS缓存提供良好的性能

## 示例输出

```
=== gRPC DNS客户端演示 ===

1. 连接到本地服务 (localhost:9090)
成功连接到 localhost:9090
连接状态: CONNECTING
健康检查失败: rpc error: code = Unavailable desc = connection error

2. 连接到外部服务 (grpc.example.com:443)
连接失败: context deadline exceeded

3. 连接到IPv4地址 (127.0.0.1:9090)
成功连接到 127.0.0.1:9090
连接状态: CONNECTING
健康检查失败: rpc error: code = Unavailable desc = connection error

=== DNS客户端演示完成 ===
```

## 与服务发现的对比

| 特性 | DNS解析器 | 服务发现 |
|------|-----------|----------|
| 配置复杂度 | 简单 | 复杂 |
| 基础设施依赖 | DNS服务器 | etcd/Consul等 |
| 动态服务发现 | 有限 | 完整支持 |
| 负载均衡 | DNS轮询 | 多种策略 |
| 健康检查 | 无 | 支持 |
| 适用场景 | 简单部署 | 微服务架构 |

## 注意事项

1. **DNS缓存**: DNS解析结果会被缓存，可能导致服务更新延迟
2. **负载均衡**: DNS轮询的负载均衡能力有限
3. **健康检查**: DNS解析器不提供健康检查功能
4. **服务发现**: 对于复杂的微服务架构，建议使用专门的服务发现解决方案

## 最佳实践

1. **开发环境**: 使用DNS解析器进行快速开发和测试
2. **生产环境**: 根据架构复杂度选择合适的解决方案
3. **混合使用**: 可以在同一应用中同时使用DNS解析器和服务发现
4. **配置管理**: 通过配置文件灵活切换解析方式