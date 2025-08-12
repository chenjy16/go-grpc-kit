package client

import (
	"context"
	"fmt"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

// discoveryResolverBuilder 服务发现解析器构建器
type discoveryResolverBuilder struct {
	serviceName string
	registry    discovery.Registry
	logger      *zap.Logger
}

// Build 构建解析器
func (b *discoveryResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &discoveryResolver{
		serviceName: b.serviceName,
		registry:    b.registry,
		logger:      b.logger,
		cc:          cc,
		ctx:         context.Background(),
	}
	
	// 启动解析器
	go r.start()
	
	return r, nil
}

// Scheme 返回解析器方案
func (b *discoveryResolverBuilder) Scheme() string {
	return "discovery"
}

// discoveryResolver 服务发现解析器
type discoveryResolver struct {
	serviceName string
	registry    discovery.Registry
	logger      *zap.Logger
	cc          resolver.ClientConn
	ctx         context.Context
	cancel      context.CancelFunc
}

// start 启动解析器
func (r *discoveryResolver) start() {
	r.ctx, r.cancel = context.WithCancel(context.Background())
	
	// 监听服务变化
	ch, err := r.registry.Watch(r.ctx, r.serviceName)
	if err != nil {
		r.logger.Error("Failed to watch services", 
			zap.String("service", r.serviceName),
			zap.Error(err))
		return
	}
	
	for {
		select {
		case services, ok := <-ch:
			if !ok {
				r.logger.Info("Service watch channel closed",
					zap.String("service", r.serviceName))
				return
			}
			
			r.updateAddresses(services)
			
		case <-r.ctx.Done():
			return
		}
	}
}

// updateAddresses 更新地址列表
func (r *discoveryResolver) updateAddresses(services []*discovery.ServiceInfo) {
	var addrs []resolver.Address
	
	for _, service := range services {
		addr := resolver.Address{
			Addr: fmt.Sprintf("%s:%d", service.Address, service.Port),
		}
		
		addrs = append(addrs, addr)
	}
	
	state := resolver.State{
		Addresses: addrs,
	}
	
	if err := r.cc.UpdateState(state); err != nil {
		r.logger.Error("Failed to update resolver state",
			zap.String("service", r.serviceName),
			zap.Error(err))
	} else {
		r.logger.Debug("Updated resolver addresses",
			zap.String("service", r.serviceName),
			zap.Int("count", len(addrs)))
	}
}

// ResolveNow 立即解析
func (r *discoveryResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	// 触发立即解析
	go func() {
		services, err := r.registry.Discover(r.ctx, r.serviceName)
		if err != nil {
			r.logger.Error("Failed to discover services",
				zap.String("service", r.serviceName),
				zap.Error(err))
			return
		}
		
		r.updateAddresses(services)
	}()
}

// Close 关闭解析器
func (r *discoveryResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}