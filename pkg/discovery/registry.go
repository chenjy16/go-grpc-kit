package discovery

import (
	"context"
	"fmt"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"go.uber.org/zap"
)

// Registry 服务注册接口
type Registry interface {
	Register(ctx context.Context, service *ServiceInfo) error
	Deregister(ctx context.Context, service *ServiceInfo) error
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)
	Close() error
}

// NewRegistry 创建服务注册器
func NewRegistry(cfg *config.DiscoveryConfig, logger *zap.Logger) (Registry, error) {
	switch cfg.Type {
	case "etcd":
		return NewEtcdRegistry(cfg.Endpoints, cfg.Namespace, logger)
	case "consul":
		return NewConsulRegistry(cfg.Endpoints, cfg.Namespace, logger)
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", cfg.Type)
	}
}

// ServiceManager 服务管理器
type ServiceManager struct {
	registry Registry
	logger   *zap.Logger
	services map[string]*ServiceInfo
}

// NewServiceManager 创建服务管理器
func NewServiceManager(registry Registry, logger *zap.Logger) *ServiceManager {
	return &ServiceManager{
		registry: registry,
		logger:   logger,
		services: make(map[string]*ServiceInfo),
	}
}

// RegisterService 注册服务
func (sm *ServiceManager) RegisterService(ctx context.Context, service *ServiceInfo) error {
	if err := sm.registry.Register(ctx, service); err != nil {
		return err
	}
	
	key := fmt.Sprintf("%s:%s:%d", service.Name, service.Address, service.Port)
	sm.services[key] = service
	
	return nil
}

// DeregisterService 注销服务
func (sm *ServiceManager) DeregisterService(ctx context.Context, service *ServiceInfo) error {
	if err := sm.registry.Deregister(ctx, service); err != nil {
		return err
	}
	
	key := fmt.Sprintf("%s:%s:%d", service.Name, service.Address, service.Port)
	delete(sm.services, key)
	
	return nil
}

// DeregisterAll 注销所有服务
func (sm *ServiceManager) DeregisterAll(ctx context.Context) error {
	for _, service := range sm.services {
		if err := sm.registry.Deregister(ctx, service); err != nil {
			sm.logger.Error("Failed to deregister service",
				zap.String("service", service.Name),
				zap.Error(err))
		}
	}
	
	sm.services = make(map[string]*ServiceInfo)
	return nil
}

// DiscoverServices 发现服务
func (sm *ServiceManager) DiscoverServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	return sm.registry.Discover(ctx, serviceName)
}

// WatchServices 监听服务变化
func (sm *ServiceManager) WatchServices(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	return sm.registry.Watch(ctx, serviceName)
}

// Close 关闭服务管理器
func (sm *ServiceManager) Close() error {
	return sm.registry.Close()
}