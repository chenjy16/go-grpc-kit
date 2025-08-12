.PHONY: build test clean proto deps example

# 构建项目
build:
	go build -o bin/grpc-kit ./cmd/...

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/

# 生成 protobuf 文件
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		examples/simple/proto/*.proto

# 安装依赖
deps:
	go mod tidy
	go mod download

# 运行示例
example:
	go run examples/simple/main.go

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 安装工具
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 初始化项目
init: install-tools deps proto

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build        - Build the project"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build files"
	@echo "  proto        - Generate protobuf files"
	@echo "  deps         - Install dependencies"
	@echo "  example      - Run example"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  install-tools - Install required tools"
	@echo "  init         - Initialize project"
	@echo "  help         - Show this help"