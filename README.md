# Go gRPC Kit

A Spring Boot-like gRPC framework for Go that provides out-of-the-box gRPC service development experience.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Core Components](#core-components)
- [Examples](#examples)
- [Project Structure](#project-structure)
- [Best Practices](#best-practices)
- [License](#license)

## Features

- 🔧 **Auto Configuration** - Automatic configuration loading from files or environment variables
- 🚀 **gRPC Server Auto Registration & Startup** - Automatic gRPC server startup with multi-service registration
- 🔍 **Service Discovery** - Support for etcd/consul auto registration, deregistration, and client resolution
- 🔒 **TLS/mTLS Support** - Built-in secure communication capabilities
- 🔗 **Interceptor Chain** - Middleware chain processing for logging, metrics collection, error recovery
- 📊 **Health/Metrics/Management Endpoints** - Health checks, Prometheus metrics exposure, management pages
- 🏭 **Client Factory** - Service name-based dialing with load balancing, timeout, and retry policies
- 🎯 **Graceful Startup & Shutdown** - Support for smooth online and offline operations
- 🌐 **DNS Resolver Support** - Built-in DNS resolver for direct connections and Nginx proxies
- ⚙️ **Comprehensive Configuration** - Extensive gRPC server and client configuration options

## Quick Start

### Method 1: Using Starter Framework (Recommended)

The simplest way to start, zero-configuration out-of-the-box:

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
    "github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
    "google.golang.org/grpc"
)

// Implement your gRPC service
type GreeterService struct {
    proto.UnimplementedGreeterServer
}

func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
    return &proto.HelloResponse{
        Message: "Hello " + req.Name + "!",
    }, nil
}

func main() {
    // Start a complete gRPC service with one line of code
    err := starter.RunSimple(func(s grpc.ServiceRegistrar) {
        proto.RegisterGreeterServer(s, &GreeterService{})
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### Method 2: Using Traditional Application Framework

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

// Implement ServiceRegistrar interface
func (s *GreeterService) RegisterService(server grpc.ServiceRegistrar) {
    proto.RegisterGreeterServer(server, s)
}

func main() {
    application := app.New()
    
    // Register service
    application.RegisterService(&GreeterService{})
    
    // Start application
    application.Run()
}
```

### Method 3: Chain Configuration

```go
package main

import (
    "github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
    "google.golang.org/grpc"
)

func main() {
    err := starter.NewStarter(
        starter.WithGrpcPort(9090),        // gRPC port
        starter.WithMetricsPort(8081),     // Metrics port
        starter.WithAppMetrics(true),      // Enable metrics
        starter.WithAppDiscovery(false),   // Disable service discovery
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

## Configuration

Create a `config/application.yml` configuration file:

```yaml
server:
  host: "0.0.0.0"
  port: 8080          # HTTP port
  grpc_port: 9090     # gRPC port

grpc:
  server:
    # Message size limits
    max_recv_msg_size: 8388608  # 8MB
    max_send_msg_size: 8388608  # 8MB
    
    # Connection configuration
    max_concurrent_streams: 200
    connection_timeout: 180     # 3 minutes
    keepalive_time: 60         # 1 minute
    keepalive_timeout: 10      # 10 seconds
    keepalive_min_time: 10     # 10 seconds
    
    # Feature switches
    enable_reflection: true
    enable_compression: true
    compression_level: "gzip"
    
    # Interceptor configuration
    enable_logging: true
    enable_metrics: true
    enable_recovery: true
    enable_tracing: false

  client:
    # Basic configuration
    timeout: 60                # 1 minute
    load_balancing: "round_robin"
    max_recv_msg_size: 8388608
    max_send_msg_size: 8388608
    
    # Keepalive configuration
    keepalive_time: 30
    keepalive_timeout: 5
    permit_without_stream: true
    
    # Retry policy
    retry_policy:
      max_attempts: 5
      initial_backoff: "1s"
      max_backoff: "60s"
      backoff_multiplier: 2.0
      retryable_status_codes:
        - "UNAVAILABLE"
        - "DEADLINE_EXCEEDED"
        - "RESOURCE_EXHAUSTED"
    
    # Compression configuration
    compression: "gzip"
    
    # Interceptor configuration
    enable_logging: true
    enable_metrics: true
    enable_tracing: false

discovery:
  type: "etcd"        # Support etcd or consul
  endpoints:
    - "localhost:2379"
  namespace: "/grpc-kit"

logging:
  level: "info"       # debug, info, warn, error
  format: "json"      # json or console

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

## Core Components

### 1. Starter Framework

Provides Spring Boot-like zero-configuration startup experience:

```go
// Simplest startup method
starter.RunSimple(func(s grpc.ServiceRegistrar) {
    proto.RegisterGreeterServer(s, &GreeterService{})
})

// Startup with configuration
starter.Run(
    starter.WithGrpcPort(9090),
    starter.WithMetricsPort(8081),
    starter.WithAppDiscovery("etcd", []string{"localhost:2379"}),
    starter.RegisterService(func(s grpc.ServiceRegistrar) {
        proto.RegisterGreeterServer(s, &GreeterService{})
    }),
)
```

#### Features

- 🚀 **Out-of-the-box**: Zero-configuration gRPC service startup
- 🔧 **Modular Design**: Independent functional modules, selectively enable/disable
- ⚙️ **Simplified Configuration**: Chain calls, Spring Boot-like configuration experience
- 📊 **Built-in Monitoring**: Auto-integrated Prometheus metrics and health checks
- 🔍 **Service Discovery**: Optional etcd service registration and discovery
- 📝 **Structured Logging**: High-performance logging based on zap
- 🛡️ **Graceful Shutdown**: Support for signal handling and graceful shutdown

#### Configuration Options

**Basic Configuration:**
- `WithGrpcPort(port int)`: Set gRPC service port (default: 9090)
- `WithMetricsPort(port int)`: Set metrics service port (default: 8081)
- `WithConfig(cfg *config.Config)`: Use custom configuration
- `WithAppLogger(logger *zap.Logger)`: Use custom logger

**Feature Switches:**
- `WithAppMetrics(enabled bool)`: Enable/disable Prometheus metrics (default: true)
- `WithAppDiscovery(enabled bool)`: Enable/disable service discovery (default: false)
- `WithEtcdEndpoints(endpoints []string)`: Set etcd endpoints

### 2. Service Discovery

Support for etcd and consul automatic service registration and discovery:

```go
// Create etcd registry
registry, err := discovery.NewEtcdRegistry([]string{"localhost:2379"})
if err != nil {
    log.Fatal(err)
}

// Register service
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

// Discover services
services, err := registry.Discover(context.Background(), "greeter-service")
if err != nil {
    log.Fatal(err)
}
```

### 3. Client Factory

Create gRPC client connections by service name:

```go
// Create client factory
factory := client.NewFactory(
    client.WithDiscovery(registry),
    client.WithTimeout(30*time.Second),
    client.WithMaxRetries(3),
)

// Get service client
conn, err := factory.GetClient("greeter-service")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Use client
greeterClient := proto.NewGreeterClient(conn)
response, err := greeterClient.SayHello(context.Background(), &proto.HelloRequest{
    Name: "World",
})
```

### 4. DNS Resolver Support

When service discovery is not configured, the framework automatically uses gRPC's built-in DNS resolver:

#### Configuration for DNS Resolution

```yaml
# Note: No discovery section configured, client will use gRPC built-in DNS resolver
grpc:
  client:
    timeout: 30
    load_balancing: "round_robin"
    # ... other client configurations
```

#### Supported Address Formats

- **Domain names**: `example.com:9090`
- **IP addresses**: `192.168.1.100:9090`
- **Local addresses**: `localhost:9090` or `127.0.0.1:9090`
- **IPv6 addresses**: `[::1]:9090`

#### Usage Example

```go
// Create application
app := app.New(app.WithConfig(cfg))

// Get client connection (using DNS resolution)
conn, err := app.GetClient("example.com:9090")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Use connection to create gRPC client
client := your_proto.NewYourServiceClient(conn)
```

### 5. Nginx Integration

Configure gRPC clients to connect to Nginx addresses:

#### Configuration

```yaml
# Note: No discovery section configured to enable DNS resolver
grpc:
  client:
    timeout: 30
    load_balancing: "round_robin"
    # ... other client configurations
```

#### Supported Nginx Address Formats

- **Domain addresses**: `nginx.example.com:443`
- **IP addresses**: `192.168.1.100:80`
- **Local addresses**: `localhost:8080`
- **Internal load balancers**: `nginx-lb.internal:9090`

#### Usage Example

```go
// Connect to Nginx address
client, err := application.GetClient("nginx.example.com:443")
if err != nil {
    log.Printf("Connection failed: %v", err)
    return
}

// Use client to call gRPC service
// userClient := pb.NewUserServiceClient(client)
// response, err := userClient.GetUser(ctx, &pb.GetUserRequest{...})
```

#### Nginx Configuration Example

```nginx
upstream grpc_backend {
    server backend1.example.com:9090;
    server backend2.example.com:9090;
    server backend3.example.com:9090;
}

server {
    listen 443 ssl http2;
    server_name nginx.example.com;
    
    # SSL configuration
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

### 6. Interceptors

Built-in multiple interceptors with chain call support:

```go
// Metrics collection interceptor
metricsInterceptor := interceptor.NewMetricsUnaryInterceptor()

// Logging interceptor
loggingInterceptor := interceptor.NewLoggingUnaryInterceptor()

// Recovery interceptor
recoveryInterceptor := interceptor.NewRecoveryUnaryInterceptor()

// Add interceptors when creating server
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recoveryInterceptor,
        loggingInterceptor,
        metricsInterceptor,
    ),
)
```

### 7. Health Checks and Metrics

Automatically provide health checks and Prometheus metrics:

```bash
# Health check
curl http://localhost:8080/health

# Prometheus metrics
curl http://localhost:8081/metrics
```

#### Built-in Metrics

- `grpc_requests_total`: Total gRPC requests
- `grpc_request_duration_seconds`: gRPC request duration
- `grpc_active_requests`: Current active requests

### 8. TLS Support

Support for TLS and mTLS secure communication:

```yaml
tls:
  enabled: true
  cert_file: "server.crt"
  key_file: "server.key"
  ca_file: "ca.crt"  # Required for mTLS
```

## Examples

The framework provides comprehensive examples in the `examples/` directory:

### 1. DNS Client Demo (`examples/dns_client_demo/`)

Demonstrates how to use the built-in DNS resolver functionality:

```bash
cd examples/dns_client_demo
go run main.go
```

**Features:**
- DNS resolver support when no service discovery is configured
- Flexible target addresses supporting domains, IPs, and ports
- Configuration-driven client behavior
- Automatic connection lifecycle management

### 2. Nginx Client Demo (`examples/nginx_client_demo/`)

Shows how to configure gRPC clients to connect to Nginx addresses:

```bash
cd examples/nginx_client_demo
go run main.go
```

**Features:**
- Direct connection to Nginx addresses
- Support for various Nginx address formats
- Comprehensive client configuration options
- Examples for connecting to upstream services through Nginx

### 3. gRPC Configuration Demo (`examples/grpc_config_demo/`)

Demonstrates enhanced gRPC configuration capabilities:

```bash
cd examples/grpc_config_demo
go run main.go
```

**Features:**
- Comprehensive server and client configuration options
- Message size limits, connection management, keepalive parameters
- Retry policies, compression support, interceptor configuration
- Real-time configuration display

## Project Structure

```
go-grpc-kit/
├── cmd/                    # Command-line tools and examples
├── config/                 # Configuration files
├── internal/              # Internal packages
├── pkg/                   # Public packages
│   ├── app/              # Application framework
│   ├── config/           # Configuration management
│   ├── discovery/        # Service discovery (etcd/consul)
│   ├── interceptor/      # Interceptors (metrics/logging/recovery)
│   ├── client/           # Client factory
│   ├── server/           # gRPC server
│   └── starter/          # Starter framework
└── examples/             # Example code
    ├── simple/           # Simple examples
    ├── discovery/        # Service discovery examples
    ├── client/           # Client examples
    ├── dns_client_demo/  # DNS client demonstration
    ├── nginx_client_demo/ # Nginx client demonstration
    └── grpc_config_demo/ # gRPC configuration demonstration
```

## Best Practices

### 1. Development Environment

- Use DNS resolver for rapid development and testing
- Leverage the starter framework for quick prototyping
- Utilize comprehensive configuration options for fine-tuning

### 2. Production Environment

- Choose appropriate solutions based on architecture complexity
- Use service discovery for complex microservice architectures
- Implement proper TLS/mTLS for secure communication
- Configure appropriate retry policies and timeouts

### 3. Configuration Management

- Use configuration files to flexibly switch resolution methods
- Implement environment-specific configurations
- Monitor and adjust performance-related settings

### 4. Monitoring and Observability

- Enable built-in metrics and health checks
- Implement structured logging for better debugging
- Use interceptors for comprehensive request tracking

### 5. Comparison: DNS Resolver vs Service Discovery

| Feature | DNS Resolver | Service Discovery |
|---------|--------------|-------------------|
| Configuration Complexity | Simple | Complex |
| Infrastructure Dependencies | DNS servers | etcd/Consul etc. |
| Dynamic Service Discovery | Limited | Full support |
| Load Balancing | DNS round-robin | Multiple strategies |
| Health Checks | None | Supported |
| Use Cases | Simple deployments | Microservice architectures |

### 6. Framework Comparison: Traditional vs Starter

#### Traditional Approach
```go
// Manual server creation, middleware configuration, monitoring startup, etc.
server := grpc.NewServer(
    grpc.UnaryInterceptor(/* various interceptors */),
)
RegisterMyServiceServer(server, &MyService{})

lis, _ := net.Listen("tcp", ":9090")
go server.Serve(lis)

// Manual metrics server startup
http.Handle("/metrics", promhttp.Handler())
go http.ListenAndServe(":8081", nil)

// Manual signal handling...
```

#### Starter Approach
```go
// Start complete gRPC service with one line of code
starter.RunSimple(func(s grpc.ServiceRegistrar) {
    RegisterMyServiceServer(s, &MyService{})
})
```

## License

MIT License