package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/go-grpc-kit/go-grpc-kit/pkg/config"
	"github.com/go-grpc-kit/go-grpc-kit/pkg/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server gRPC 服务器
type Server struct {
	config     *config.Config
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
	services   []ServiceRegistrar
	mu         sync.RWMutex
	started    bool
	healthSrv  *health.Server
}

// ServiceRegistrar 服务注册接口
type ServiceRegistrar interface {
	RegisterService(s grpc.ServiceRegistrar)
}

// New 创建新的 gRPC 服务器
func New(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		services:  make([]ServiceRegistrar, 0),
		healthSrv: health.NewServer(),
	}
}

// RegisterService 注册服务
func (s *Server) RegisterService(service ServiceRegistrar) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.started {
		s.logger.Warn("Cannot register service after server started")
		return
	}
	
	s.services = append(s.services, service)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.started {
		return fmt.Errorf("server already started")
	}
	
	// 创建监听器
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener
	
	// 创建 gRPC 服务器选项
	opts, err := s.buildServerOptions()
	if err != nil {
		return fmt.Errorf("failed to build server options: %w", err)
	}
	
	// 创建 gRPC 服务器
	s.grpcServer = grpc.NewServer(opts...)
	
	// 注册健康检查服务
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthSrv)
	
	// 注册反射服务（开发环境）
	reflection.Register(s.grpcServer)
	
	// 注册业务服务
	for _, service := range s.services {
		service.RegisterService(s.grpcServer)
	}
	
	s.started = true
	
	s.logger.Info("gRPC server starting", 
		zap.String("address", addr),
		zap.Int("services", len(s.services)))
	
	// 启动服务器
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			s.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()
	
	// 设置健康状态
	s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	
	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.started {
		return nil
	}
	
	s.logger.Info("Stopping gRPC server...")
	
	// 设置健康状态为不可用
	s.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	
	// 优雅关闭
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	
	// 等待优雅关闭或超时
	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn("Force stopping gRPC server due to timeout")
		s.grpcServer.Stop()
	}
	
	s.started = false
	return nil
}

// buildServerOptions 构建服务器选项
func (s *Server) buildServerOptions() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption
	
	// 设置消息大小限制
	opts = append(opts,
		grpc.MaxRecvMsgSize(s.config.GRPC.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(s.config.GRPC.Server.MaxSendMsgSize),
	)
	
	// 添加拦截器链
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptor.LoggingUnaryInterceptor(s.logger),
		interceptor.RecoveryUnaryInterceptor(s.logger),
		interceptor.MetricsUnaryInterceptor(),
	}
	
	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptor.LoggingStreamInterceptor(s.logger),
		interceptor.RecoveryStreamInterceptor(s.logger),
		interceptor.MetricsStreamInterceptor(),
	}
	
	opts = append(opts,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)
	
	// TLS 配置
	if s.config.TLS.Enabled {
		creds, err := s.buildTLSCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}
	
	return opts, nil
}

// buildTLSCredentials 构建 TLS 凭证
func (s *Server) buildTLSCredentials() (credentials.TransportCredentials, error) {
	if s.config.TLS.CertFile == "" || s.config.TLS.KeyFile == "" {
		return nil, fmt.Errorf("TLS cert file and key file must be specified")
	}
	
	cert, err := tls.LoadX509KeyPair(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}
	
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
	}
	
	// 如果配置了 CA 文件，启用 mTLS
	if s.config.TLS.CAFile != "" {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		// TODO: 加载 CA 证书
	}
	
	return credentials.NewTLS(tlsConfig), nil
}

// GetAddress 获取服务器地址
func (s *Server) GetAddress() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// IsHealthy 检查服务器健康状态
func (s *Server) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// SetHealthStatus 设置服务健康状态
func (s *Server) SetHealthStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.healthSrv.SetServingStatus(service, status)
}