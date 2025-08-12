package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ConsulRegistry consul 服务注册器
type ConsulRegistry struct {
	client    *api.Client
	logger    *zap.Logger
	namespace string
}

// NewConsulRegistry 创建 consul 注册器
func NewConsulRegistry(endpoints []string, namespace string, logger *zap.Logger) (*ConsulRegistry, error) {
	config := api.DefaultConfig()
	if len(endpoints) > 0 {
		config.Address = endpoints[0]
	}
	
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	
	return &ConsulRegistry{
		client:    client,
		logger:    logger,
		namespace: namespace,
	}, nil
}

// Register 注册服务
func (r *ConsulRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	serviceID := fmt.Sprintf("%s-%s-%d", service.Name, service.Address, service.Port)
	
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    []string{"grpc"},
		Meta:    service.Metadata,
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", service.Address, service.Port),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}
	
	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	
	r.logger.Info("Service registered to consul",
		zap.String("service", service.Name),
		zap.String("address", service.Address),
		zap.Int("port", service.Port),
		zap.String("service_id", serviceID))
	
	return nil
}

// Deregister 注销服务
func (r *ConsulRegistry) Deregister(ctx context.Context, service *ServiceInfo) error {
	serviceID := fmt.Sprintf("%s-%s-%d", service.Name, service.Address, service.Port)
	
	if err := r.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	
	r.logger.Info("Service deregistered from consul",
		zap.String("service", service.Name),
		zap.String("service_id", serviceID))
	
	return nil
}

// Discover 发现服务
func (r *ConsulRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	services, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}
	
	var result []*ServiceInfo
	for _, service := range services {
		info := &ServiceInfo{
			Name:     service.Service.Service,
			Address:  service.Service.Address,
			Port:     service.Service.Port,
			Metadata: service.Service.Meta,
		}
		result = append(result, info)
	}
	
	return result, nil
}

// Watch 监听服务变化
func (r *ConsulRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	ch := make(chan []*ServiceInfo, 1)
	
	// 首次获取服务列表
	services, err := r.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	ch <- services
	
	// 启动监听协程
	go func() {
		defer close(ch)
		
		var lastIndex uint64
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			
			// 使用 blocking query 监听变化
			queryOpts := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			}
			
			services, meta, err := r.client.Health().Service(serviceName, "", true, queryOpts)
			if err != nil {
				r.logger.Error("Failed to watch services", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
			
			lastIndex = meta.LastIndex
			
			var result []*ServiceInfo
			for _, service := range services {
				info := &ServiceInfo{
					Name:     service.Service.Service,
					Address:  service.Service.Address,
					Port:     service.Service.Port,
					Metadata: service.Service.Meta,
				}
				result = append(result, info)
			}
			
			select {
			case ch <- result:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch, nil
}

// Close 关闭注册器
func (r *ConsulRegistry) Close() error {
	// Consul client 不需要显式关闭
	return nil
}