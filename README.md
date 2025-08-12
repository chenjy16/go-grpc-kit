# Go gRPC Kit

一个类似 Spring Boot 的 Go gRPC 框架，提供开箱即用的 gRPC 服务开发体验。

## 核心功能

- 🔧 **自动配置** - 基于配置文件或环境变量自动加载配置
- 🚀 **gRPC Server 自动注册与启动** - 自动启动 gRPC 服务端，支持多服务注册
- 🔍 **服务发现** - 支持 etcd/consul 的自动注册、注销和客户端解析
- 🔒 **TLS/mTLS 支持** - 内置安全通信能力
- 🔗 **拦截器链** - 日志、指标收集、错误恢复等中间件链式处理
- 📊 **Health/Metrics/管理端口** - 健康检查、Prometheus 指标暴露、管理页面
- 🏭 **客户端工厂** - 按服务名 dial，支持负载均衡、超时、重试策略
- 🎯 **优雅启动与关闭** - 支持平滑上线和下线

## 快速开始

### 方式一：使用 Starter 框架（推荐）

最简单的启动方式，零配置开箱即用：

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
    "github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
    "google.golang.org/grpc"
)

// 实现你的 gRPC 服务
type GreeterService struct {
    proto.UnimplementedGreeterServer
}

func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
    return &proto.HelloResponse{
        Message: "Hello " + req.Name + "!",
    }, nil
}

func main() {
    // 一行代码启动完整的 gRPC 服务
    err := starter.RunSimple(func(s grpc.ServiceRegistrar) {
        proto.RegisterGreeterServer(s, &GreeterService{})
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### 方式二：使用传统 Application 框架

```go
package main

import (
    "context"
    "github.com/go-grpc-kit/go-grpc-kit/pkg/app"
    "github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
    "google.golang.org/grpc"
)

type GreeterService struct {
    proto.UnimplementedGreeterServer
}

func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
    return &proto.HelloResponse{Message: "Hello " + req.Name}, nil
}

// 实现 ServiceRegistrar 接口
func (s *GreeterService) RegisterService(server grpc.ServiceRegistrar) {
    proto.RegisterGreeterServer(server, s)
}

func main() {
    application := app.New()
    
    // 注册服务
    application.RegisterService(&GreeterService{})
    
    // 启动应用
    application.Run()
}
```

## 配置文件

创建 `config/application.yml` 配置文件：

```yaml
server:
  host: "0.0.0.0"
  port: 8080          # HTTP 端口
  grpc_port: 9090     # gRPC 端口

grpc:
  server:
    max_recv_msg_size: 4194304  # 4MB
    max_send_msg_size: 4194304  # 4MB
  client:
    timeout: 30
    max_retries: 3
    load_balancing: "round_robin"

discovery:
  type: "etcd"        # 支持 etcd 或 consul
  endpoints:
    - "localhost:2379"
  namespace: "/grpc-kit"

logging:
  level: "info"       # debug, info, warn, error
  format: "json"      # json 或 console

tls:
  enabled: false
  cert_file: ""
  key_file: ""
  ca_file: ""

metrics:
  enabled: true
  port: 8081
  path: "/metrics"
```

## 核心功能详解

### 1. Starter 框架

提供类似 Spring Boot 的零配置启动体验：

```go
// 最简单的启动方式
starter.RunSimple(func(s grpc.ServiceRegistrar) {
    proto.RegisterGreeterServer(s, &GreeterService{})
})

// 带配置的启动方式
starter.Run(
    starter.WithGrpcPort(9090),
    starter.WithMetricsPort(8081),
    starter.WithAppDiscovery("etcd", []string{"localhost:2379"}),
    starter.RegisterService(func(s grpc.ServiceRegistrar) {
        proto.RegisterGreeterServer(s, &GreeterService{})
    }),
)
```

### 2. 服务发现

支持 etcd 和 consul 的自动服务注册与发现：

```go
// 创建 etcd 注册中心
registry, err := discovery.NewEtcdRegistry([]string{"localhost:2379"})
if err != nil {
    log.Fatal(err)
}

// 注册服务
serviceInfo := &discovery.ServiceInfo{
    Name:    "greeter-service",
    Version: "v1.0.0",
    Address: "localhost:9090",
    Metadata: map[string]string{
        "protocol": "grpc",
    },
}

err = registry.Register(context.Background(), serviceInfo)
if err != nil {
    log.Fatal(err)
}

// 发现服务
services, err := registry.Discover(context.Background(), "greeter-service")
if err != nil {
    log.Fatal(err)
}
```

### 3. 客户端工厂

按服务名创建 gRPC 客户端连接：

```go
// 创建客户端工厂
factory := client.NewFactory(
    client.WithDiscovery(registry),
    client.WithTimeout(30*time.Second),
    client.WithMaxRetries(3),
)

// 获取服务客户端
conn, err := factory.GetClient("greeter-service")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// 使用客户端
greeterClient := proto.NewGreeterClient(conn)
response, err := greeterClient.SayHello(context.Background(), &proto.HelloRequest{
    Name: "World",
})
```

### 4. 拦截器

内置多种拦截器，支持链式调用：

```go
// 指标收集拦截器
metricsInterceptor := interceptor.NewMetricsUnaryInterceptor()

// 日志拦截器
loggingInterceptor := interceptor.NewLoggingUnaryInterceptor()

// 恢复拦截器
recoveryInterceptor := interceptor.NewRecoveryUnaryInterceptor()

// 创建服务器时添加拦截器
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recoveryInterceptor,
        loggingInterceptor,
        metricsInterceptor,
    ),
)
```

### 5. 健康检查和指标

自动提供健康检查和 Prometheus 指标：

```bash
# 健康检查
curl http://localhost:8080/health

# Prometheus 指标
curl http://localhost:8081/metrics
```

### 6. TLS 支持

支持 TLS 和 mTLS 安全通信：

```yaml
tls:
  enabled: true
  cert_file: "server.crt"
  key_file: "server.key"
  ca_file: "ca.crt"  # mTLS 需要
```

## 项目结构

```
go-grpc-kit/
├── cmd/                    # 命令行工具和示例
├── config/                 # 配置文件
├── internal/              # 内部包
├── pkg/                   # 公共包
│   ├── app/              # 应用框架
│   ├── config/           # 配置管理
│   ├── discovery/        # 服务发现（etcd/consul）
│   ├── interceptor/      # 拦截器（指标/日志/恢复）
│   ├── client/           # 客户端工厂
│   ├── server/           # gRPC 服务端
│   └── starter/          # Starter 框架
└── examples/             # 示例代码
    ├── simple/           # 简单示例
    ├── discovery/        # 服务发现示例
    └── client/           # 客户端示例
```

## 示例

查看 `examples/` 目录获取更多使用示例：

- `examples/simple/` - 基础 gRPC 服务示例
- `examples/discovery/` - 服务发现示例
- `examples/client/` - 客户端使用示例

## 许可证

MIT License