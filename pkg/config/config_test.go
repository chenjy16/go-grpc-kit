package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 测试加载默认配置
	cfg := Get()
	if cfg == nil {
		t.Fatal("Expected config to be loaded")
	}

	// 验证默认值
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.GRPCPort != 9090 {
		t.Errorf("Expected gRPC port to be 9090, got %d", cfg.Server.GRPCPort)
	}

	if cfg.Discovery.Type != "etcd" {
		t.Errorf("Expected discovery type to be etcd, got %s", cfg.Discovery.Type)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("GRPC_KIT_SERVER_PORT", "8888")
	os.Setenv("GRPC_KIT_DISCOVERY_TYPE", "consul")
	defer func() {
		os.Unsetenv("GRPC_KIT_SERVER_PORT")
		os.Unsetenv("GRPC_KIT_DISCOVERY_TYPE")
	}()

	// 重新加载配置
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证环境变量覆盖
	if cfg.Server.Port != 8888 {
		t.Errorf("Expected server port to be 8888, got %d", cfg.Server.Port)
	}

	if cfg.Discovery.Type != "consul" {
		t.Errorf("Expected discovery type to be consul, got %s", cfg.Discovery.Type)
	}
}

func TestGetEnv(t *testing.T) {
	// 测试存在的环境变量
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := GetEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected test_value, got %s", value)
	}

	// 测试不存在的环境变量
	value = GetEnv("NON_EXISTENT_VAR", "default")
	if value != "default" {
		t.Errorf("Expected default, got %s", value)
	}
}