package main

import (
	"context"
	"log"

	"github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
	"google.golang.org/grpc"
)

// GreeterService 实现 Greeter 服务
type GreeterService struct {
	proto.UnimplementedGreeterServer
}

// SayHello 实现 SayHello 方法
func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{
		Message: "Hello " + req.Name + " from advanced starter!",
	}, nil
}

func main() {
	log.Println("Starting advanced gRPC service with custom configuration...")

	// 方式2: 使用链式调用和自定义配置
	err := starter.NewStarter(
		starter.WithGrpcPort(9091),        // 自定义 gRPC 端口
		starter.WithMetricsPort(8082),     // 自定义指标端口
		starter.WithAppMetrics(true),      // 启用指标
		starter.WithAppDiscovery(false),   // 禁用服务发现
	).
		AddService(starter.NewSimpleService(func(s grpc.ServiceRegistrar) {
			proto.RegisterGreeterServer(s, &GreeterService{})
		})).
		Run()

	if err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}