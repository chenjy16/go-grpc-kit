package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/discovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

// ClientFactory gRPC 客户端工厂
type ClientFactory struct {
	config    *config.Config
	logger    *zap.Logger
	registry  discovery.Registry
	clients   map[string]*grpc.ClientConn
	mu        sync.RWMutex
}

// NewClientFactory 创建客户端工厂
func NewClientFactory(cfg *config.Config, registry discovery.Registry, logger *zap.Logger) *ClientFactory {
	return &ClientFactory{
		config:   cfg,
		logger:   logger,
		registry: registry,
		clients:  make(map[string]*grpc.ClientConn),
	}
}

// GetClient 获取客户端连接
func (f *ClientFactory) GetClient(serviceName string) (*grpc.ClientConn, error) {
	f.mu.RLock()
	if conn, exists := f.clients[serviceName]; exists {
		f.mu.RUnlock()
		return conn, nil
	}
	f.mu.RUnlock()
	
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 双重检查
	if conn, exists := f.clients[serviceName]; exists {
		return conn, nil
	}
	
	// 创建新连接
	conn, err := f.createConnection(serviceName)
	if err != nil {
		return nil, err
	}
	
	f.clients[serviceName] = conn
	return conn, nil
}

// createConnection 创建连接
func (f *ClientFactory) createConnection(serviceName string) (*grpc.ClientConn, error) {
	// 首先检查服务是否存在
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	services, err := f.registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}
	
	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	
	// 构建连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(f.buildServiceConfig()),
	}
	
	// 设置消息大小限制
	if f.config.GRPC.Client.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(f.config.GRPC.Client.MaxRecvMsgSize)))
	}
	if f.config.GRPC.Client.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(f.config.GRPC.Client.MaxSendMsgSize)))
	}
	
	// 设置 Keepalive 配置
	if f.config.GRPC.Client.KeepaliveTime > 0 {
		keepaliveParams := keepalive.ClientParameters{
			Time:                time.Duration(f.config.GRPC.Client.KeepaliveTime) * time.Second,
			Timeout:             time.Duration(f.config.GRPC.Client.KeepaliveTimeout) * time.Second,
			PermitWithoutStream: f.config.GRPC.Client.PermitWithoutStream,
		}
		opts = append(opts, grpc.WithKeepaliveParams(keepaliveParams))
	}
	
	// 添加拦截器
	opts = append(opts, f.buildInterceptors()...)
	
	// 确定目标地址
	var target string
	if f.registry != nil {
		// 使用服务发现解析器
		target = fmt.Sprintf("discovery:///%s", serviceName)
		// 注册自定义解析器
		f.registerResolver(serviceName)
	} else {
		// 直接使用DNS解析，serviceName应该是host:port格式
		target = serviceName
		f.logger.Info("Using DNS resolver for gRPC client",
			zap.String("service", serviceName),
			zap.String("target", target))
	}
	
	// 创建连接
	ctx2, cancel2 := context.WithTimeout(context.Background(), 
		time.Duration(f.config.GRPC.Client.Timeout)*time.Second)
	defer cancel2()
	
	conn, err := grpc.DialContext(ctx2, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", serviceName, err)
	}
	
	f.logger.Info("Created gRPC client connection",
		zap.String("service", serviceName),
		zap.String("target", target))
	
	return conn, nil
}

// buildServiceConfig 构建服务配置
func (f *ClientFactory) buildServiceConfig() string {
	retryPolicy := f.config.GRPC.Client.RetryPolicy
	
	// 构建重试状态码数组
	statusCodes := "["
	for i, code := range retryPolicy.RetryableStatusCodes {
		if i > 0 {
			statusCodes += ", "
		}
		statusCodes += fmt.Sprintf(`"%s"`, code)
	}
	statusCodes += "]"
	
	return fmt.Sprintf(`{
		"loadBalancingPolicy": "%s",
		"retryPolicy": {
			"maxAttempts": %d,
			"initialBackoff": "%s",
			"maxBackoff": "%s",
			"backoffMultiplier": %f,
			"retryableStatusCodes": %s
		}
	}`, f.config.GRPC.Client.LoadBalancing, 
		retryPolicy.MaxAttempts,
		retryPolicy.InitialBackoff,
		retryPolicy.MaxBackoff,
		retryPolicy.BackoffMultiplier,
		statusCodes)
}

// buildInterceptors 构建拦截器
func (f *ClientFactory) buildInterceptors() []grpc.DialOption {
	var opts []grpc.DialOption
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor
	
	// 根据配置添加拦截器
	if f.config.GRPC.Client.EnableLogging {
		unaryInterceptors = append(unaryInterceptors, f.loggingUnaryInterceptor())
		streamInterceptors = append(streamInterceptors, f.loggingStreamInterceptor())
	}
	
	if f.config.GRPC.Client.EnableMetrics {
		unaryInterceptors = append(unaryInterceptors, f.metricsUnaryInterceptor())
		streamInterceptors = append(streamInterceptors, f.metricsStreamInterceptor())
	}
	
	// TODO: 添加 tracing 拦截器支持
	// if f.config.GRPC.Client.EnableTracing {
	//     unaryInterceptors = append(unaryInterceptors, f.tracingUnaryInterceptor())
	//     streamInterceptors = append(streamInterceptors, f.tracingStreamInterceptor())
	// }
	
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptors...))
	}
	
	return opts
}

// registerResolver 注册自定义解析器
func (f *ClientFactory) registerResolver(serviceName string) {
	builder := &discoveryResolverBuilder{
		serviceName: serviceName,
		registry:    f.registry,
		logger:      f.logger,
	}
	resolver.Register(builder)
}

// Close 关闭所有客户端连接
func (f *ClientFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	for serviceName, conn := range f.clients {
		if err := conn.Close(); err != nil {
			f.logger.Error("Failed to close client connection",
				zap.String("service", serviceName),
				zap.Error(err))
		}
	}
	
	f.clients = make(map[string]*grpc.ClientConn)
	return nil
}

// loggingUnaryInterceptor 客户端一元调用日志拦截器
func (f *ClientFactory) loggingUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)
		
		if err != nil {
			f.logger.Error("gRPC client call failed",
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			f.logger.Debug("gRPC client call completed",
				zap.String("method", method),
				zap.Duration("duration", duration))
		}
		
		return err
	}
}

// loggingStreamInterceptor 客户端流式调用日志拦截器
func (f *ClientFactory) loggingStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, opts...)
		duration := time.Since(start)
		
		if err != nil {
			f.logger.Error("gRPC client stream failed",
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Error(err))
		} else {
			f.logger.Debug("gRPC client stream created",
				zap.String("method", method),
				zap.Duration("duration", duration))
		}
		
		return stream, err
	}
}

// metricsUnaryInterceptor 客户端一元调用指标拦截器
func (f *ClientFactory) metricsUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// TODO: 实现客户端指标收集
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// metricsStreamInterceptor 客户端流式调用指标拦截器
func (f *ClientFactory) metricsStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// TODO: 实现客户端指标收集
		return streamer(ctx, desc, cc, method, opts...)
	}
}