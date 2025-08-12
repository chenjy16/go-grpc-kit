package interceptor

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// gRPC 请求总数
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "code"},
	)
	
	// gRPC 请求持续时间
	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
	
	// gRPC 当前活跃请求数
	grpcActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_active_requests",
			Help: "Number of active gRPC requests",
		},
		[]string{"method"},
	)
)

// MetricsUnaryInterceptor 一元调用指标拦截器
func MetricsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod
		
		// 增加活跃请求数
		grpcActiveRequests.WithLabelValues(method).Inc()
		defer grpcActiveRequests.WithLabelValues(method).Dec()
		
		// 调用处理器
		resp, err := handler(ctx, req)
		
		// 记录指标
		duration := time.Since(start).Seconds()
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		
		codeStr := strconv.Itoa(int(code))
		grpcRequestsTotal.WithLabelValues(method, codeStr).Inc()
		grpcRequestDuration.WithLabelValues(method, codeStr).Observe(duration)
		
		return resp, err
	}
}

// MetricsStreamInterceptor 流式调用指标拦截器
func MetricsStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		method := info.FullMethod
		
		// 增加活跃请求数
		grpcActiveRequests.WithLabelValues(method).Inc()
		defer grpcActiveRequests.WithLabelValues(method).Dec()
		
		// 调用处理器
		err := handler(srv, stream)
		
		// 记录指标
		duration := time.Since(start).Seconds()
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		
		codeStr := strconv.Itoa(int(code))
		grpcRequestsTotal.WithLabelValues(method, codeStr).Inc()
		grpcRequestDuration.WithLabelValues(method, codeStr).Observe(duration)
		
		return err
	}
}

// GetMetricsRegistry 获取指标注册表
func GetMetricsRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}