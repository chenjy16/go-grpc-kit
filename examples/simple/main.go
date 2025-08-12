package main

import (
	"context"
	"log"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/app"
	"github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
	"google.golang.org/grpc"
)

// GreeterService 示例服务
type GreeterService struct {
	proto.UnimplementedGreeterServer
}

// SayHello 实现 SayHello 方法
func (s *GreeterService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{
		Message: "Hello " + req.Name + "!",
	}, nil
}

// RegisterService 实现 ServiceRegistrar 接口
func (s *GreeterService) RegisterService(server grpc.ServiceRegistrar) {
	proto.RegisterGreeterServer(server, s)
}

func main() {
	// 创建应用程序
	application := app.New()
	
	// 注册服务
	application.RegisterService(&GreeterService{})
	
	// 启动应用程序
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}