package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/app"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
)

var (
	configFile = flag.String("config", "", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	Version = "1.0.0"
	Name    = "go-grpc-kit"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", Name, Version)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建应用程序
	application := app.New(
		app.WithConfig(cfg),
	)

	// 启动应用程序
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}