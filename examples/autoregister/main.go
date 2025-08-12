package main

import (
	"log"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/starter"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./config/application.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 启用自动注册
	cfg.AutoRegister.Enabled = true
	cfg.AutoRegister.ScanDirs = []string{"./services"}

	// 创建并运行应用
	app := starter.New(
		starter.WithConfig(cfg),
	)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}