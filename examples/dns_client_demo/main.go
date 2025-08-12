package main

import (
	"context"
	"log"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/app"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./config/application.yml")
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = config.Get()
	}

	// 创建应用程序
	application := app.New(
		app.WithConfig(cfg),
	)

	// 在单独的goroutine中运行应用程序
	go func() {
		if err := application.Run(); err != nil {
			log.Printf("Application failed: %v", err)
		}
	}()

	// 等待应用程序初始化
	time.Sleep(2 * time.Second)

	// 演示DNS客户端连接
	demonstrateDNSClient(application)
}

func demonstrateDNSClient(app *app.Application) {
	log.Println("=== gRPC DNS客户端演示 ===")

	// 示例1: 连接到本地服务（使用DNS解析）
	log.Println("\n1. 连接到本地服务 (localhost:9090)")
	conn1, err := app.GetClient("localhost:9090")
	if err != nil {
		log.Printf("连接失败: %v", err)
	} else {
		log.Printf("成功连接到 localhost:9090")
		testConnection(conn1, "localhost:9090")
		conn1.Close()
	}

	// 示例2: 连接到外部服务（使用DNS解析）
	log.Println("\n2. 连接到外部服务 (grpc.example.com:443)")
	conn2, err := app.GetClient("grpc.example.com:443")
	if err != nil {
		log.Printf("连接失败: %v", err)
	} else {
		log.Printf("成功连接到 grpc.example.com:443")
		testConnection(conn2, "grpc.example.com:443")
		conn2.Close()
	}

	// 示例3: 连接到IPv4地址
	log.Println("\n3. 连接到IPv4地址 (127.0.0.1:9090)")
	conn3, err := app.GetClient("127.0.0.1:9090")
	if err != nil {
		log.Printf("连接失败: %v", err)
	} else {
		log.Printf("成功连接到 127.0.0.1:9090")
		testConnection(conn3, "127.0.0.1:9090")
		conn3.Close()
	}

	log.Println("\n=== DNS客户端演示完成 ===")
}

func testConnection(conn *grpc.ClientConn, target string) {
	// 测试连接状态
	state := conn.GetState()
	log.Printf("连接状态: %v", state)

	// 尝试健康检查（如果服务支持）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		log.Printf("健康检查失败: %v", err)
	} else {
		log.Printf("健康检查成功: %v", resp.Status)
	}
}