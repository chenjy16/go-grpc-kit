package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// EtcdRegistry etcd 服务注册器
type EtcdRegistry struct {
	client    *clientv3.Client
	logger    *zap.Logger
	namespace string
	ttl       int64
	leaseID   clientv3.LeaseID
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewEtcdRegistry 创建 etcd 注册器
func NewEtcdRegistry(endpoints []string, namespace string, logger *zap.Logger) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	_, err = client.Status(ctx, endpoints[0])
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}
	
	return &EtcdRegistry{
		client:    client,
		logger:    logger,
		namespace: namespace,
		ttl:       30, // 30 秒 TTL
	}, nil
}

// Register 注册服务
func (r *EtcdRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	// 创建租约
	lease, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = lease.ID
	
	// 序列化服务信息
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}
	
	// 构建服务键
	key := r.buildServiceKey(service.Name, service.Address, service.Port)
	
	// 注册服务
	_, err = r.client.Put(ctx, key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	
	// 启动租约续期
	ch, kaerr := r.client.KeepAlive(ctx, lease.ID)
	if kaerr != nil {
		return fmt.Errorf("failed to keep alive lease: %w", kaerr)
	}
	
	// 处理续期响应
	go func() {
		for ka := range ch {
			r.logger.Debug("Lease renewed", zap.Int64("lease_id", int64(ka.ID)))
		}
	}()
	
	r.logger.Info("Service registered",
		zap.String("service", service.Name),
		zap.String("address", service.Address),
		zap.Int("port", service.Port),
		zap.String("key", key))
	
	return nil
}

// Deregister 注销服务
func (r *EtcdRegistry) Deregister(ctx context.Context, service *ServiceInfo) error {
	// 撤销租约
	if r.leaseID != 0 {
		_, err := r.client.Revoke(ctx, r.leaseID)
		if err != nil {
			r.logger.Warn("Failed to revoke lease", zap.Error(err))
		}
	}
	
	// 删除服务键
	key := r.buildServiceKey(service.Name, service.Address, service.Port)
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	
	r.logger.Info("Service deregistered",
		zap.String("service", service.Name),
		zap.String("address", service.Address),
		zap.Int("port", service.Port))
	
	return nil
}

// Discover 发现服务
func (r *EtcdRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	prefix := r.buildServicePrefix(serviceName)
	
	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}
	
	var services []*ServiceInfo
	for _, kv := range resp.Kvs {
		var service ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			r.logger.Warn("Failed to unmarshal service info",
				zap.String("key", string(kv.Key)),
				zap.Error(err))
			continue
		}
		services = append(services, &service)
	}
	
	return services, nil
}

// Watch 监听服务变化
func (r *EtcdRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	prefix := r.buildServicePrefix(serviceName)
	ch := make(chan []*ServiceInfo, 1)
	
	// 首次获取服务列表
	services, err := r.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	ch <- services
	
	// 监听变化
	watchCh := r.client.Watch(ctx, prefix, clientv3.WithPrefix())
	
	go func() {
		defer close(ch)
		for watchResp := range watchCh {
			if watchResp.Err() != nil {
				r.logger.Error("Watch error", zap.Error(watchResp.Err()))
				return
			}
			
			// 重新获取服务列表
			services, err := r.Discover(ctx, serviceName)
			if err != nil {
				r.logger.Error("Failed to discover services on watch", zap.Error(err))
				continue
			}
			
			select {
			case ch <- services:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch, nil
}

// Close 关闭注册器
func (r *EtcdRegistry) Close() error {
	return r.client.Close()
}

// buildServiceKey 构建服务键
func (r *EtcdRegistry) buildServiceKey(serviceName, address string, port int) string {
	return path.Join(r.namespace, "services", serviceName, fmt.Sprintf("%s:%d", address, port))
}

// buildServicePrefix 构建服务前缀
func (r *EtcdRegistry) buildServicePrefix(serviceName string) string {
	return path.Join(r.namespace, "services", serviceName) + "/"
}