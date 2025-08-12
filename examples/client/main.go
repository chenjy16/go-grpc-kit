package main

import (
	"context"
	"log"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/client"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Get()
	}

	// 创建日志器
	logger, _ := zap.NewDevelopment()

	// 创建服务发现注册器
	registry, err := discovery.NewRegistry(&cfg.Discovery, logger)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	defer registry.Close()

	// 创建客户端工厂
	clientFactory := client.NewClientFactory(cfg, registry, logger)
	defer clientFactory.Close()

	// 获取客户端连接
	conn, err := clientFactory.GetClient("greeter")
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	// 创建 gRPC 客户端
	client := proto.NewGreeterClient(conn)

	// 调用服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &proto.HelloRequest{
		Name: "World",
	})
	if err != nil {
		log.Fatalf("Failed to call SayHello: %v", err)
	}

	log.Printf("Response: %s", resp.Message)
}