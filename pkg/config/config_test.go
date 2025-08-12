package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestDefaultGRPCServerConfig(t *testing.T) {
	config, err := Load("")
	assert.NoError(t, err)
	
	// 测试默认的服务器配置
	assert.Equal(t, 4*1024*1024, config.GRPC.Server.MaxRecvMsgSize)
	assert.Equal(t, 4*1024*1024, config.GRPC.Server.MaxSendMsgSize)
	assert.Equal(t, uint32(100), config.GRPC.Server.MaxConcurrentStreams)
	assert.Equal(t, 120, config.GRPC.Server.ConnectionTimeout)
	assert.Equal(t, 30, config.GRPC.Server.KeepaliveTime)
	assert.Equal(t, 5, config.GRPC.Server.KeepaliveTimeout)
	assert.Equal(t, 5, config.GRPC.Server.KeepaliveMinTime)
	assert.False(t, config.GRPC.Server.EnableReflection)
	assert.False(t, config.GRPC.Server.EnableCompression)
	assert.Equal(t, "gzip", config.GRPC.Server.CompressionLevel)
	assert.True(t, config.GRPC.Server.EnableLogging)
	assert.True(t, config.GRPC.Server.EnableMetrics)
	assert.True(t, config.GRPC.Server.EnableRecovery)
	assert.False(t, config.GRPC.Server.EnableTracing)
}

func TestDefaultGRPCClientConfig(t *testing.T) {
	config, err := Load("")
	assert.NoError(t, err)
	
	// 测试默认的客户端配置
	assert.Equal(t, 4*1024*1024, config.GRPC.Client.MaxRecvMsgSize)
	assert.Equal(t, 4*1024*1024, config.GRPC.Client.MaxSendMsgSize)
	assert.Equal(t, 30, config.GRPC.Client.Timeout)
	assert.Equal(t, 30, config.GRPC.Client.KeepaliveTime)
	assert.Equal(t, 5, config.GRPC.Client.KeepaliveTimeout)
	assert.False(t, config.GRPC.Client.PermitWithoutStream)
	assert.Equal(t, "round_robin", config.GRPC.Client.LoadBalancing)
	assert.False(t, config.GRPC.Client.EnableCompression)
	assert.Equal(t, "gzip", config.GRPC.Client.CompressionLevel)
	assert.True(t, config.GRPC.Client.EnableLogging)
	assert.True(t, config.GRPC.Client.EnableMetrics)
	assert.False(t, config.GRPC.Client.EnableTracing)
}

func TestDefaultRetryPolicyConfig(t *testing.T) {
	config, err := Load("")
	assert.NoError(t, err)
	
	// 测试默认的重试策略配置
	retryPolicy := config.GRPC.Client.RetryPolicy
	assert.Equal(t, 3, retryPolicy.MaxAttempts)
	assert.Equal(t, "1s", retryPolicy.InitialBackoff)
	assert.Equal(t, "30s", retryPolicy.MaxBackoff)
	assert.Equal(t, 2.0, retryPolicy.BackoffMultiplier)
	assert.Contains(t, retryPolicy.RetryableStatusCodes, "UNAVAILABLE")
	assert.Contains(t, retryPolicy.RetryableStatusCodes, "DEADLINE_EXCEEDED")
}

func TestRetryPolicyDurations(t *testing.T) {
	config, err := Load("")
	assert.NoError(t, err)
	retryPolicy := config.GRPC.Client.RetryPolicy
	
	// 测试退避时间解析
	initialBackoff, err := time.ParseDuration(retryPolicy.InitialBackoff)
	assert.NoError(t, err)
	assert.Equal(t, time.Second, initialBackoff)
	
	maxBackoff, err := time.ParseDuration(retryPolicy.MaxBackoff)
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Second, maxBackoff)
	
	// 确保最大退避时间大于初始退避时间
	assert.Greater(t, maxBackoff, initialBackoff)
}