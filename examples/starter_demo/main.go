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
		Message: "Hello " + req.Name + "!",
	}, nil
}

func main() {
	// 方式1: 最简单的启动方式
	log.Println("Starting gRPC service with starter framework...")
	
	err := starter.RunSimple(func(s grpc.ServiceRegistrar) {
		proto.RegisterGreeterServer(s, &GreeterService{})
	}, starter.DefaultOptions()...)
	
	if err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}