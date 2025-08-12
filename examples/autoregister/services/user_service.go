package services

import (
	"context"

	"github.com/go-grpc-kit/go-grpc-kit/examples/simple/proto"
	"google.golang.org/grpc"
)

// UserService 用户服务
// @grpc-service UserService
type UserService struct {
	proto.UnimplementedGreeterServer
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{}
}

// SayHello 实现 Greeter 服务的 SayHello 方法
func (s *UserService) SayHello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{
		Message: "Hello from UserService: " + req.Name,
	}, nil
}

// RegisterService 注册服务到 gRPC 服务器
func (s *UserService) RegisterService(server grpc.ServiceRegistrar) {
	proto.RegisterGreeterServer(server, s)
}