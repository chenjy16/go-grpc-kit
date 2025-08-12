# gRPC Starter Framework

一个类似 Spring Boot 的 gRPC 微服务启动框架，提供开箱即用的功能和简化的配置。

## 特性

- 🚀 **开箱即用**: 零配置启动 gRPC 服务
- 🔧 **模块化设计**: 功能模块独立，可选择性启用/禁用
- ⚙️ **简化配置**: 链式调用，类似 Spring Boot 的配置体验
- 📊 **内置监控**: 自动集成 Prometheus 指标和健康检查
- 🔍 **服务发现**: 可选的 etcd 服务注册与发现
- 📝 **结构化日志**: 基于 zap 的高性能日志
- 🛡️ **优雅关闭**: 支持信号处理和优雅关闭

## 快速开始

### 最简单的方式

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
    "google.golang.org/grpc"
)

// 实现你的 gRPC 服务
type MyService struct {
    // 你的服务实现
}

func main() {
    err := starter.RunSimple(func(s grpc.ServiceRegistrar) {
        // 注册你的服务
        RegisterMyServiceServer(s, &MyService{})
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### 链式配置方式

```go
package main

import (
    "github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
    "google.golang.org/grpc"
)

func main() {
    err := starter.NewStarter(
        starter.WithGrpcPort(9090),        // gRPC 端口
        starter.WithMetricsPort(8081),     // 指标端口
        starter.WithAppMetrics(true),      // 启用指标
        starter.WithAppDiscovery(false),   // 禁用服务发现
    ).
        AddService(starter.NewSimpleService(func(s grpc.ServiceRegistrar) {
            RegisterMyServiceServer(s, &MyService{})
        })).
        Run()
        
    if err != nil {
        log.Fatal(err)
    }
}
```

### 多服务注册

```go
func main() {
    services := []starter.ServiceRegistrar{
        starter.NewSimpleService(func(s grpc.ServiceRegistrar) {
            RegisterService1Server(s, &Service1{})
        }),
        starter.NewSimpleService(func(s grpc.ServiceRegistrar) {
            RegisterService2Server(s, &Service2{})
        }),
    }
    
    err := starter.RunWithServices(services, starter.DefaultOptions()...)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 配置选项

### 基础配置

- `WithGrpcPort(port int)`: 设置 gRPC 服务端口（默认: 9090）
- `WithMetricsPort(port int)`: 设置指标服务端口（默认: 8081）
- `WithConfig(cfg *config.Config)`: 使用自定义配置
- `WithAppLogger(logger *zap.Logger)`: 使用自定义日志器

### 功能开关

- `WithAppMetrics(enabled bool)`: 启用/禁用 Prometheus 指标（默认: true）
- `WithAppDiscovery(enabled bool)`: 启用/禁用服务发现（默认: false）
- `WithEtcdEndpoints(endpoints []string)`: 设置 etcd 端点

### 默认配置

```go
starter.DefaultOptions() // 返回以下默认配置:
// - gRPC 端口: 9090
// - 指标端口: 8081
// - 启用指标: true
// - 禁用服务发现: false
```

## 内置功能

### 1. Prometheus 指标

自动暴露以下指标：

- `grpc_requests_total`: gRPC 请求总数
- `grpc_request_duration_seconds`: gRPC 请求持续时间
- `grpc_active_requests`: 当前活跃请求数

访问地址: `http://localhost:8081/metrics`

### 2. 健康检查

自动提供健康检查端点：

访问地址: `http://localhost:8081/health`

### 3. gRPC 反射

默认启用 gRPC 反射，支持工具如 grpcurl：

```bash
grpcurl -plaintext localhost:9090 list
```

### 4. 结构化日志

基于 zap 的高性能结构化日志，支持 JSON 格式输出。

### 5. 优雅关闭

自动处理 SIGINT 和 SIGTERM 信号，确保服务优雅关闭。

## 模块化架构

框架采用模块化设计，每个功能都是独立的模块：

- **GrpcServerModule**: gRPC 服务器核心模块
- **MetricsModule**: Prometheus 指标和健康检查模块
- **DiscoveryModule**: 服务注册与发现模块

你可以通过配置选择性地启用或禁用这些模块。

## 示例

查看 `examples/` 目录下的完整示例：

- `examples/starter_demo/`: 基础使用示例
- `examples/starter_advanced/`: 高级配置示例

## 与传统方式对比

### 传统方式
```go
// 需要手动创建服务器、配置中间件、启动监控等
server := grpc.NewServer(
    grpc.UnaryInterceptor(/* 各种拦截器 */),
)
RegisterMyServiceServer(server, &MyService{})

lis, _ := net.Listen("tcp", ":9090")
go server.Serve(lis)

// 手动启动指标服务器
http.Handle("/metrics", promhttp.Handler())
go http.ListenAndServe(":8081", nil)

// 手动处理信号...
```

### Starter 方式
```go
// 一行代码启动完整的 gRPC 服务
starter.RunSimple(func(s grpc.ServiceRegistrar) {
    RegisterMyServiceServer(s, &MyService{})
})
```

## 配置文件支持

框架支持 YAML 配置文件，放置在 `config/application.yaml` 或当前目录下：

```yaml
server:
  grpc_port: 9090
  host: "0.0.0.0"

metrics:
  enabled: true
  port: 8081
  path: "/metrics"

discovery:
  type: "etcd"
  endpoints:
    - "localhost:2379"
  namespace: "/grpc-kit"

logging:
  level: "info"
  format: "json"
```

## 环境变量

所有配置都可以通过环境变量覆盖，使用 `GRPC_KIT_` 前缀：

```bash
export GRPC_KIT_SERVER_GRPC_PORT=9091
export GRPC_KIT_METRICS_ENABLED=false
```