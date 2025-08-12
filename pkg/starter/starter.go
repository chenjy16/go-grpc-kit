package starter

import (
	"google.golang.org/grpc"
)

// GrpcStarter 简化的 gRPC 启动器
type GrpcStarter struct {
	app *GrpcApplication
}

// NewStarter 创建新的启动器
func NewStarter(opts ...AppOption) *GrpcStarter {
	return &GrpcStarter{
		app: New(opts...),
	}
}

// AddService 添加服务（链式调用）
func (s *GrpcStarter) AddService(service ServiceRegistrar) *GrpcStarter {
	s.app.RegisterService(service)
	return s
}

// AddModule 添加模块（链式调用）
func (s *GrpcStarter) AddModule(module Module) *GrpcStarter {
	s.app.RegisterModule(module)
	return s
}

// Run 运行应用
func (s *GrpcStarter) Run() error {
	return s.app.Run()
}

// SimpleService 简单服务包装器
type SimpleService struct {
	registerFunc func(grpc.ServiceRegistrar)
}

// NewSimpleService 创建简单服务
func NewSimpleService(registerFunc func(grpc.ServiceRegistrar)) ServiceRegistrar {
	return &SimpleService{
		registerFunc: registerFunc,
	}
}

// RegisterService 实现 ServiceRegistrar 接口
func (s *SimpleService) RegisterService(server grpc.ServiceRegistrar) {
	s.registerFunc(server)
}

// 便捷函数

// RunWithService 运行带有单个服务的应用
func RunWithService(service ServiceRegistrar, opts ...AppOption) error {
	return NewStarter(opts...).
		AddService(service).
		Run()
}

// RunWithServices 运行带有多个服务的应用
func RunWithServices(services []ServiceRegistrar, opts ...AppOption) error {
	starter := NewStarter(opts...)
	for _, service := range services {
		starter.AddService(service)
	}
	return starter.Run()
}

// RunSimple 最简单的运行方式
func RunSimple(registerFunc func(grpc.ServiceRegistrar), opts ...AppOption) error {
	service := NewSimpleService(registerFunc)
	return RunWithService(service, opts...)
}