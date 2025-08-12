package autoregister

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	PackageName string
	TypeName    string
	FilePath    string
	ServiceName string
}

// Scanner 服务扫描器
type Scanner struct {
	config *config.AutoRegisterConfig
	logger *zap.Logger
	fset   *token.FileSet
}

// NewScanner 创建新的扫描器
func NewScanner(cfg *config.AutoRegisterConfig, logger *zap.Logger) *Scanner {
	return &Scanner{
		config: cfg,
		logger: logger,
		fset:   token.NewFileSet(),
	}
}

// ScanServices 扫描服务
func (s *Scanner) ScanServices() ([]*ServiceInfo, error) {
	var services []*ServiceInfo

	for _, dir := range s.config.ScanDirs {
		dirServices, err := s.scanDirectory(dir)
		if err != nil {
			s.logger.Warn("Failed to scan directory", 
				zap.String("dir", dir), 
				zap.Error(err))
			continue
		}
		services = append(services, dirServices...)
	}

	return services, nil
}

// scanDirectory 扫描目录
func (s *Scanner) scanDirectory(dir string) ([]*ServiceInfo, error) {
	var services []*ServiceInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过非 Go 文件
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 检查是否匹配模式
		if !s.matchesPattern(path) {
			return nil
		}

		// 检查是否被排除
		if s.isExcluded(path) {
			return nil
		}

		fileServices, err := s.scanFile(path)
		if err != nil {
			s.logger.Warn("Failed to scan file", 
				zap.String("file", path), 
				zap.Error(err))
			return nil
		}

		services = append(services, fileServices...)
		return nil
	})

	return services, err
}

// scanFile 扫描文件
func (s *Scanner) scanFile(filePath string) ([]*ServiceInfo, error) {
	src, err := parser.ParseFile(s.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	var services []*ServiceInfo

	ast.Inspect(src, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if s.isServiceType(node, src) {
				service := &ServiceInfo{
					PackageName: src.Name.Name,
					TypeName:    node.Name.Name,
					FilePath:    filePath,
					ServiceName: s.extractServiceName(node.Name.Name),
				}
				services = append(services, service)
			}
		}
		return true
	})

	return services, nil
}

// isServiceType 检查是否是服务类型
func (s *Scanner) isServiceType(typeSpec *ast.TypeSpec, file *ast.File) bool {
	// 检查是否实现了 ServiceRegistrar 接口
	// 这里可以通过多种方式检查：
	// 1. 检查方法签名
	// 2. 检查注释标记
	// 3. 检查命名约定

	// 方法1: 检查是否有 RegisterService 方法
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if recv, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := recv.X.(*ast.Ident); ok {
						if ident.Name == typeSpec.Name.Name && 
						   funcDecl.Name.Name == "RegisterService" {
							return true
						}
					}
				}
			}
		}
	}

	// 方法2: 检查注释标记
	if typeSpec.Comment != nil {
		for _, comment := range typeSpec.Comment.List {
			if strings.Contains(comment.Text, "@grpc-service") {
				return true
			}
		}
	}

	// 方法3: 检查命名约定（以 Service 结尾）
	if strings.HasSuffix(typeSpec.Name.Name, "Service") {
		return true
	}

	return false
}

// matchesPattern 检查文件是否匹配模式
func (s *Scanner) matchesPattern(filePath string) bool {
	if len(s.config.Patterns) == 0 {
		return true // 如果没有指定模式，匹配所有文件
	}

	for _, pattern := range s.config.Patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
	}
	return false
}

// isExcluded 检查文件是否被排除
func (s *Scanner) isExcluded(filePath string) bool {
	for _, exclude := range s.config.Excludes {
		if matched, _ := filepath.Match(exclude, filepath.Base(filePath)); matched {
			return true
		}
		// 也检查完整路径
		if strings.Contains(filePath, exclude) {
			return true
		}
	}
	return false
}

// extractServiceName 提取服务名称
func (s *Scanner) extractServiceName(typeName string) string {
	if s.config.ServiceName != "" {
		// 使用配置的服务名称模式
		return strings.ReplaceAll(s.config.ServiceName, "{type}", typeName)
	}
	
	// 默认规则：移除 Service 后缀并转换为小写
	name := strings.TrimSuffix(typeName, "Service")
	return strings.ToLower(name)
}