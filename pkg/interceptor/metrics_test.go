package interceptor

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestMetricsUnaryInterceptor(t *testing.T) {
	interceptor := MetricsUnaryInterceptor()

	if interceptor == nil {
		t.Error("Expected interceptor to be created")
	}
}

func TestMetricsStreamInterceptor(t *testing.T) {
	interceptor := MetricsStreamInterceptor()

	if interceptor == nil {
		t.Error("Expected interceptor to be created")
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	interceptor := MetricsUnaryInterceptor()
	
	// 模拟处理器
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	
	// 模拟请求信息
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}
	
	// 调用拦截器
	resp, err := interceptor(context.Background(), "request", info, handler)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if resp != "response" {
		t.Errorf("Expected response 'response', got %v", resp)
	}
}

func TestUnaryServerInterceptorWithError(t *testing.T) {
	interceptor := MetricsUnaryInterceptor()
	
	// 模拟返回错误的处理器
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.Internal, "internal error")
	}
	
	// 模拟请求信息
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}
	
	// 调用拦截器
	resp, err := interceptor(context.Background(), "request", info, handler)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
	
	// 检查错误码
	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
	}
	
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal error code, got %v", st.Code())
	}
}

func TestStreamServerInterceptor(t *testing.T) {
	interceptor := MetricsStreamInterceptor()
	
	// 模拟处理器
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}
	
	// 模拟流信息
	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/TestStream",
	}
	
	// 调用拦截器
	err := interceptor(nil, &mockServerStream{}, info, handler)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestStreamServerInterceptorWithError(t *testing.T) {
	interceptor := MetricsStreamInterceptor()
	
	// 模拟返回错误的处理器
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}
	
	// 模拟流信息
	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/TestStream",
	}
	
	// 调用拦截器
	err := interceptor(nil, &mockServerStream{}, info, handler)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	// 检查错误码
	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
	}
	
	if st.Code() != codes.InvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", st.Code())
	}
}

func TestGetMetricsRegistry(t *testing.T) {
	registry := GetMetricsRegistry()
	
	if registry == nil {
		t.Error("Expected registry to be returned")
	}
	
	// 验证是否是 Prometheus 默认注册表
	if registry != prometheus.DefaultRegisterer {
		t.Error("Expected default Prometheus registry")
	}
}

func TestConcurrentRequests(t *testing.T) {
	interceptor := MetricsUnaryInterceptor()
	
	// 模拟处理器
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	
	// 模拟请求信息
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ConcurrentTest",
	}
	
	// 并发调用拦截器
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := interceptor(context.Background(), "request", info, handler)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			done <- true
		}()
	}
	
	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// mockServerStream 模拟 gRPC ServerStream
type mockServerStream struct{}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return context.Background() }
func (m *mockServerStream) SendMsg(interface{}) error    { return nil }
func (m *mockServerStream) RecvMsg(interface{}) error    { return nil }

// BenchmarkUnaryServerInterceptor 性能测试
func BenchmarkUnaryServerInterceptor(b *testing.B) {
	interceptor := MetricsUnaryInterceptor()
	
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/BenchmarkMethod",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := interceptor(context.Background(), "request", info, handler)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkStreamServerInterceptor(b *testing.B) {
	interceptor := MetricsStreamInterceptor()
	
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}
	
	info := &grpc.StreamServerInfo{
		FullMethod: "/test.Service/BenchmarkStream",
	}
	
	stream := &mockServerStream{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := interceptor(nil, stream, info, handler)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}