package main

import (
	"context"
	"log"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/app"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"google.golang.org/grpc/connectivity"
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

	// 演示连接到不同的Nginx地址
	nginxAddresses := []string{
		"nginx.example.com:443",     // 域名地址
		"192.168.1.100:80",          // IP地址
		"localhost:8080",            // 本地地址
		"nginx-lb.internal:9090",    // 内部负载均衡器地址
	}

	for _, address := range nginxAddresses {
		log.Printf("尝试连接到Nginx地址: %s", address)
		
		// 获取客户端连接
		client, err := application.GetClient(address)
		if err != nil {
			log.Printf("连接失败 %s: %v", address, err)
			continue
		}

		// 检查连接状态
		state := client.GetState()
		log.Printf("连接状态: %s", state.String())

		// 等待连接就绪
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if client.WaitForStateChange(ctx, connectivity.Idle) {
			newState := client.GetState()
			log.Printf("连接状态变更为: %s", newState.String())
		}
		cancel()

		// 如果Nginx配置了gRPC健康检查，可以进行健康检查
		if state == connectivity.Ready {
			healthClient := grpc_health_v1.NewHealthClient(client)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			
			resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
				Service: "", // 空字符串表示检查整体服务健康状态
			})
			
			if err != nil {
				log.Printf("健康检查失败 %s: %v", address, err)
			} else {
				log.Printf("健康检查成功 %s: %s", address, resp.Status.String())
			}
			cancel()
		}

		log.Printf("成功连接到 %s\n", address)
	}

	// 演示使用特定的Nginx上游服务
	log.Println("\n=== 演示连接到Nginx上游服务 ===")
	
	// 假设Nginx配置了以下上游服务
	upstreamServices := map[string]string{
		"user-service":    "nginx-gateway.example.com:443",
		"order-service":   "nginx-gateway.example.com:443", 
		"payment-service": "nginx-gateway.example.com:443",
	}

	for serviceName, nginxAddress := range upstreamServices {
		log.Printf("连接到服务 %s 通过 Nginx: %s", serviceName, nginxAddress)
		
		client, err := application.GetClient(nginxAddress)
		if err != nil {
			log.Printf("连接失败: %v", err)
			continue
		}

		// 检查连接状态
		state := client.GetState()
		log.Printf("服务 %s 连接状态: %s", serviceName, state.String())
		
		// 这里可以使用client调用具体的gRPC服务方法
		// 例如：userClient := pb.NewUserServiceClient(client)
		//      response, err := userClient.GetUser(ctx, &pb.GetUserRequest{...})
	}

	log.Println("Nginx客户端演示完成")
}