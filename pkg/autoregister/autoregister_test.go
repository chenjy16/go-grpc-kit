package autoregister

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

func TestNewAutoRegister(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{"./services"},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	if autoReg == nil {
		t.Fatal("Expected AutoRegister to be created")
	}

	if autoReg.config != cfg {
		t.Error("Expected config to be set")
	}

	if autoReg.logger != logger {
		t.Error("Expected logger to be set")
	}

	if autoReg.scanner == nil {
		t.Error("Expected scanner to be initialized")
	}

	if autoReg.generator == nil {
		t.Error("Expected generator to be initialized")
	}
}

func TestAutoRegisterScanAndGenerate(t *testing.T) {
	// 创建临时目录和测试文件
	tempDir := t.TempDir()
	servicesDir := filepath.Join(tempDir, "services")
	err := os.MkdirAll(servicesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}

	// 创建测试服务文件
	testServiceContent := `package services

import (
	"context"
	"google.golang.org/grpc"
)

// @grpc-service TestService
type TestService struct{}

func (s *TestService) RegisterService(server grpc.ServiceRegistrar) {
	// Register service implementation
}

func (s *TestService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "Hello"}, nil
}
`

	testServicePath := filepath.Join(servicesDir, "test_service.go")
	err = os.WriteFile(testServicePath, []byte(testServiceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service file: %v", err)
	}

	// 配置AutoRegister
	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{servicesDir},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	// 执行扫描和生成
	outputPath := filepath.Join(tempDir, "auto_register_generated.go")
	err = autoReg.ScanAndGenerate(outputPath)
	if err != nil {
		t.Fatalf("Failed to scan and generate: %v", err)
	}

	// 验证输出文件是否创建
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created")
	}

	// 读取生成的文件内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// 验证生成的代码包含预期内容
	if !contains(contentStr, "TestService") {
		t.Error("Generated code should contain TestService")
	}

	if !contains(contentStr, "AutoRegisterServices") {
		t.Error("Generated code should contain AutoRegisterServices function")
	}
}

func TestAutoRegisterScanAndGenerateEmptyDirectory(t *testing.T) {
	// 创建空的临时目录
	tempDir := t.TempDir()
	emptyDir := filepath.Join(tempDir, "empty")
	err := os.MkdirAll(emptyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{emptyDir},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	outputPath := filepath.Join(tempDir, "auto_register_generated.go")
	err = autoReg.ScanAndGenerate(outputPath)
	if err != nil {
		t.Fatalf("Failed to scan and generate for empty directory: %v", err)
	}

	// 验证输出文件是否创建（即使没有服务）
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created even for empty directory")
	}
}

func TestAutoRegisterScanAndGenerateNonExistentDirectory(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{"/non/existent/directory"},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "auto_register_generated.go")
	
	// 应该能够处理不存在的目录而不出错
	err := autoReg.ScanAndGenerate(outputPath)
	if err != nil {
		t.Fatalf("Should handle non-existent directory gracefully: %v", err)
	}
}

func TestAutoRegisterScanAndGenerateMultipleDirectories(t *testing.T) {
	// 创建多个临时目录和测试文件
	tempDir := t.TempDir()
	
	// 第一个服务目录
	servicesDir1 := filepath.Join(tempDir, "services1")
	err := os.MkdirAll(servicesDir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create services1 directory: %v", err)
	}

	service1Content := `package services1

import "google.golang.org/grpc"

type UserService struct{}

func (s *UserService) RegisterService(server grpc.ServiceRegistrar) {}
`

	err = os.WriteFile(filepath.Join(servicesDir1, "user_service.go"), []byte(service1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create user service file: %v", err)
	}

	// 第二个服务目录
	servicesDir2 := filepath.Join(tempDir, "services2")
	err = os.MkdirAll(servicesDir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create services2 directory: %v", err)
	}

	service2Content := `package services2

import "google.golang.org/grpc"

// @grpc-service OrderService
type OrderService struct{}

func (s *OrderService) RegisterService(server grpc.ServiceRegistrar) {}
`

	err = os.WriteFile(filepath.Join(servicesDir2, "order_service.go"), []byte(service2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create order service file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{servicesDir1, servicesDir2},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	outputPath := filepath.Join(tempDir, "auto_register_generated.go")
	err = autoReg.ScanAndGenerate(outputPath)
	if err != nil {
		t.Fatalf("Failed to scan and generate for multiple directories: %v", err)
	}

	// 读取生成的文件内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// 验证两个服务都被包含
	if !contains(contentStr, "UserService") {
		t.Error("Generated code should contain UserService")
	}

	if !contains(contentStr, "OrderService") {
		t.Error("Generated code should contain OrderService")
	}
}

func TestAutoRegisterScanAndGenerateWithExcludes(t *testing.T) {
	// 创建临时目录和测试文件
	tempDir := t.TempDir()
	servicesDir := filepath.Join(tempDir, "services")
	err := os.MkdirAll(servicesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}

	// 创建正常服务文件
	serviceContent := `package services

import "google.golang.org/grpc"

type UserService struct{}

func (s *UserService) RegisterService(server grpc.ServiceRegistrar) {}
`

	err = os.WriteFile(filepath.Join(servicesDir, "user_service.go"), []byte(serviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create user service file: %v", err)
	}

	// 创建测试文件（应该被排除）
	testContent := `package services

import "testing"

func TestUserService(t *testing.T) {}
`

	err = os.WriteFile(filepath.Join(servicesDir, "user_service_test.go"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{servicesDir},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	outputPath := filepath.Join(tempDir, "auto_register_generated.go")
	err = autoReg.ScanAndGenerate(outputPath)
	if err != nil {
		t.Fatalf("Failed to scan and generate with excludes: %v", err)
	}

	// 读取生成的文件内容
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// 验证正常服务被包含
	if !contains(contentStr, "UserService") {
		t.Error("Generated code should contain UserService")
	}

	// 验证测试文件内容没有被包含
	if contains(contentStr, "TestUserService") {
		t.Error("Generated code should not contain test functions")
	}
}

func TestAutoRegisterScanAndGenerateInvalidOutputPath(t *testing.T) {
	tempDir := t.TempDir()
	servicesDir := filepath.Join(tempDir, "services")
	err := os.MkdirAll(servicesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{servicesDir},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	logger := zap.NewNop()

	autoReg := NewAutoRegister(cfg, logger)

	// 使用无效的输出路径
	invalidOutputPath := "/invalid/path/auto_register_generated.go"
	err = autoReg.ScanAndGenerate(invalidOutputPath)
	if err == nil {
		t.Error("Expected error when using invalid output path")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && containsAt(s, substr, 0)))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			if start+1 <= len(s)-len(substr) {
				return containsAt(s, substr, start+1)
			}
			return false
		}
	}
	return true
}