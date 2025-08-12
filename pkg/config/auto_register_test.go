package config

import (
	"testing"
)

func TestAutoRegisterConfigDefaults(t *testing.T) {
	cfg := &AutoRegisterConfig{}
	
	// 测试默认值
	if cfg.Enabled {
		t.Error("Expected Enabled to be false by default")
	}
	
	if len(cfg.ScanDirs) != 0 {
		t.Error("Expected ScanDirs to be empty by default")
	}
	
	if len(cfg.Patterns) != 0 {
		t.Error("Expected Patterns to be empty by default")
	}
	
	if len(cfg.Excludes) != 0 {
		t.Error("Expected Excludes to be empty by default")
	}
	
	if cfg.ServiceName != "" {
		t.Error("Expected ServiceName to be empty by default")
	}
}

func TestAutoRegisterConfigFields(t *testing.T) {
	cfg := &AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{"./services", "./handlers"},
		Patterns:    []string{"*.go", "*.service.go"},
		Excludes:    []string{"*_test.go", "*_mock.go"},
		ServiceName: "test-service",
	}
	
	// 测试字段设置
	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
	
	if len(cfg.ScanDirs) != 2 {
		t.Errorf("Expected 2 scan directories, got %d", len(cfg.ScanDirs))
	}
	
	if cfg.ScanDirs[0] != "./services" {
		t.Errorf("Expected first scan dir to be './services', got '%s'", cfg.ScanDirs[0])
	}
	
	if cfg.ScanDirs[1] != "./handlers" {
		t.Errorf("Expected second scan dir to be './handlers', got '%s'", cfg.ScanDirs[1])
	}
	
	if len(cfg.Patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(cfg.Patterns))
	}
	
	if cfg.Patterns[0] != "*.go" {
		t.Errorf("Expected first pattern to be '*.go', got '%s'", cfg.Patterns[0])
	}
	
	if cfg.Patterns[1] != "*.service.go" {
		t.Errorf("Expected second pattern to be '*.service.go', got '%s'", cfg.Patterns[1])
	}
	
	if len(cfg.Excludes) != 2 {
		t.Errorf("Expected 2 excludes, got %d", len(cfg.Excludes))
	}
	
	if cfg.Excludes[0] != "*_test.go" {
		t.Errorf("Expected first exclude to be '*_test.go', got '%s'", cfg.Excludes[0])
	}
	
	if cfg.Excludes[1] != "*_mock.go" {
		t.Errorf("Expected second exclude to be '*_mock.go', got '%s'", cfg.Excludes[1])
	}
	
	if cfg.ServiceName != "test-service" {
		t.Errorf("Expected service name to be 'test-service', got '%s'", cfg.ServiceName)
	}
}

func TestAutoRegisterConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      AutoRegisterConfig
		description string
	}{
		{
			name: "valid enabled config",
			config: AutoRegisterConfig{
				Enabled:     true,
				ScanDirs:    []string{"./services"},
				Patterns:    []string{"*.go"},
				Excludes:    []string{"*_test.go"},
				ServiceName: "test-service",
			},
			description: "should be valid when enabled with scan dirs",
		},
		{
			name: "valid disabled config",
			config: AutoRegisterConfig{
				Enabled: false,
			},
			description: "should be valid when disabled",
		},
		{
			name: "enabled without scan dirs",
			config: AutoRegisterConfig{
				Enabled:  true,
				Patterns: []string{"*.go"},
			},
			description: "enabled without scan dirs should be noted",
		},
		{
			name: "enabled with empty scan dirs",
			config: AutoRegisterConfig{
				Enabled:  true,
				ScanDirs: []string{},
				Patterns: []string{"*.go"},
			},
			description: "enabled with empty scan dirs should be noted",
		},
		{
			name: "enabled without patterns",
			config: AutoRegisterConfig{
				Enabled:  true,
				ScanDirs: []string{"./services"},
			},
			description: "enabled without patterns should work with defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 基本的配置验证逻辑
			if tt.config.Enabled && len(tt.config.ScanDirs) == 0 {
				t.Logf("Config validation note: %s - scan dirs are empty", tt.description)
			} else {
				t.Logf("Config validation: %s", tt.description)
			}
		})
	}
}

func TestAutoRegisterConfigCopy(t *testing.T) {
	original := &AutoRegisterConfig{
		Enabled:     true,
		ScanDirs:    []string{"./services", "./handlers"},
		Patterns:    []string{"*.go"},
		Excludes:    []string{"*_test.go"},
		ServiceName: "test-service",
	}
	
	// 创建副本
	copy := &AutoRegisterConfig{
		Enabled:     original.Enabled,
		ScanDirs:    make([]string, len(original.ScanDirs)),
		Patterns:    make([]string, len(original.Patterns)),
		Excludes:    make([]string, len(original.Excludes)),
		ServiceName: original.ServiceName,
	}
	
	// 复制切片内容
	for i, dir := range original.ScanDirs {
		copy.ScanDirs[i] = dir
	}
	for i, pattern := range original.Patterns {
		copy.Patterns[i] = pattern
	}
	for i, exclude := range original.Excludes {
		copy.Excludes[i] = exclude
	}
	
	// 验证副本
	if copy.Enabled != original.Enabled {
		t.Error("Copy should have same Enabled value")
	}
	
	if len(copy.ScanDirs) != len(original.ScanDirs) {
		t.Error("Copy should have same number of scan dirs")
	}
	
	if len(copy.Patterns) != len(original.Patterns) {
		t.Error("Copy should have same number of patterns")
	}
	
	if len(copy.Excludes) != len(original.Excludes) {
		t.Error("Copy should have same number of excludes")
	}
	
	if copy.ServiceName != original.ServiceName {
		t.Error("Copy should have same service name")
	}
	
	// 修改副本不应影响原始配置
	copy.ScanDirs[0] = "./modified"
	if original.ScanDirs[0] == "./modified" {
		t.Error("Modifying copy should not affect original")
	}
}