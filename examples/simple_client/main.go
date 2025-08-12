package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 解析命令行参数
	addr := flag.String("addr", "localhost:9090", "gRPC server address")
	flag.Parse()

	// 直接连接到 gRPC 服务器
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

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