package autoregister

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

func TestNewScanner(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{"./test"},
		Patterns: []string{"*.go"},
		Excludes: []string{"*_test.go"},
	}
	logger := zap.NewNop()

	scanner := NewScanner(cfg, logger)

	if scanner == nil {
		t.Fatal("Expected scanner to be created")
	}

	if scanner.config != cfg {
		t.Error("Expected config to be set")
	}

	if scanner.logger != logger {
		t.Error("Expected logger to be set")
	}

	if scanner.fset == nil {
		t.Error("Expected file set to be initialized")
	}
}

func TestScanServices(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	// 创建测试服务文件
	serviceContent := `package services

import (
	"context"
	"google.golang.org/grpc"
)

// TestService 测试服务
// @grpc-service TestService
type TestService struct{}

// RegisterService 注册服务
func (s *TestService) RegisterService(server grpc.ServiceRegistrar) {
	// 注册逻辑
}

// SayHello 测试方法
func (s *TestService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "Hello"}, nil
}
`

	serviceFile := filepath.Join(tempDir, "test_service.go")
	err := os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service file: %v", err)
	}

	// 创建非服务文件
	nonServiceContent := `package services

type Helper struct{}

func (h *Helper) Help() string {
	return "help"
}
`

	nonServiceFile := filepath.Join(tempDir, "helper.go")
	err = os.WriteFile(nonServiceFile, []byte(nonServiceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-service file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{tempDir},
		Patterns: []string{"*.go"},
		Excludes: []string{"*_test.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if len(services) > 0 {
		service := services[0]
		if service.TypeName != "TestService" {
			t.Errorf("Expected service name 'TestService', got '%s'", service.TypeName)
		}
		if service.PackageName != "services" {
			t.Errorf("Expected package name 'services', got '%s'", service.PackageName)
		}
		if service.ServiceName != "test" {
			t.Errorf("Expected service name 'test', got '%s'", service.ServiceName)
		}
	}
}

func TestScanServicesEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{tempDir},
		Patterns: []string{"*.go"},
		Excludes: []string{"*_test.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	if len(services) != 0 {
		t.Errorf("Expected 0 services in empty directory, got %d", len(services))
	}
}

func TestScanServicesNonExistentDirectory(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{"/non/existent/directory"},
		Patterns: []string{"*.go"},
		Excludes: []string{"*_test.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	// 应该返回空列表而不是错误
	if len(services) != 0 {
		t.Errorf("Expected 0 services for non-existent directory, got %d", len(services))
	}
}

func TestMatchesPattern(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		Patterns: []string{"*.go", "service_*.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.go", true},
		{"service_user.go", true},
		{"test.txt", false},
		{"service.py", false},
	}

	for _, test := range tests {
		result := scanner.matchesPattern(test.path)
		if result != test.expected {
			t.Errorf("matchesPattern(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	cfg := &config.AutoRegisterConfig{
		Excludes: []string{"*_test.go", "*_mock.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	tests := []struct {
		path     string
		expected bool
	}{
		{"service_test.go", true},
		{"user_mock.go", true},
		{"service.go", false},
		{"user.go", false},
	}

	for _, test := range tests {
		result := scanner.isExcluded(test.path)
		if result != test.expected {
			t.Errorf("isExcluded(%s) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

func TestIsServiceTypeWithComment(t *testing.T) {
	tempDir := t.TempDir()

	// 创建带有注释标记的服务文件
	serviceContent := `package services

// UserService 用户服务
// @grpc-service UserService
type UserService struct{}
`

	serviceFile := filepath.Join(tempDir, "user_service.go")
	err := os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{tempDir},
		Patterns: []string{"*.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service with comment annotation, got %d", len(services))
	}
}

func TestIsServiceTypeWithRegisterMethod(t *testing.T) {
	tempDir := t.TempDir()

	// 创建带有 RegisterService 方法的服务文件
	serviceContent := `package services

import "google.golang.org/grpc"

type OrderService struct{}

func (s *OrderService) RegisterService(server grpc.ServiceRegistrar) {
	// 注册逻辑
}
`

	serviceFile := filepath.Join(tempDir, "order_service.go")
	err := os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{tempDir},
		Patterns: []string{"*.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service with RegisterService method, got %d", len(services))
	}
}

func TestIsServiceTypeWithNamingConvention(t *testing.T) {
	tempDir := t.TempDir()

	// 创建符合命名约定的服务文件
	serviceContent := `package services

type PaymentService struct{}
`

	serviceFile := filepath.Join(tempDir, "payment_service.go")
	err := os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test service file: %v", err)
	}

	cfg := &config.AutoRegisterConfig{
		ScanDirs: []string{tempDir},
		Patterns: []string{"*.go"},
	}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	services, err := scanner.ScanServices()
	if err != nil {
		t.Fatalf("Failed to scan services: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service with naming convention, got %d", len(services))
	}
}

func TestExtractServiceName(t *testing.T) {
	cfg := &config.AutoRegisterConfig{}
	logger := zap.NewNop()
	scanner := NewScanner(cfg, logger)

	tests := []struct {
		typeName string
		expected string
	}{
		{"UserService", "user"},
		{"OrderService", "order"},
		{"PaymentGatewayService", "paymentgateway"},
		{"Service", ""},
		{"TestService", "test"},
	}

	for _, test := range tests {
		result := scanner.extractServiceName(test.typeName)
		if result != test.expected {
			t.Errorf("extractServiceName(%s) = %s, expected %s", test.typeName, result, test.expected)
		}
	}
}

func TestExtractServiceNameWithPattern(t *testing.T) {
	config := &config.AutoRegisterConfig{
		ServiceName: "custom-{type}",
	}
	scanner := NewScanner(config, zap.NewNop())

	result := scanner.extractServiceName("UserService")
	expected := "custom-UserService"
	if result != expected {
		t.Errorf("extractServiceName with pattern = %s, expected %s", result, expected)
	}
}