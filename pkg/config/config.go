package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server       ServerConfig       `mapstructure:"server" yaml:"server"`
	GRPC         GRPCConfig         `mapstructure:"grpc" yaml:"grpc"`
	Discovery    DiscoveryConfig    `mapstructure:"discovery" yaml:"discovery"`
	Logging      LoggingConfig      `mapstructure:"logging" yaml:"logging"`
	TLS          TLSConfig          `mapstructure:"tls" yaml:"tls"`
	Metrics      MetricsConfig      `mapstructure:"metrics" yaml:"metrics"`
	AutoRegister AutoRegisterConfig `mapstructure:"auto_register" yaml:"auto_register"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port     int    `mapstructure:"port" yaml:"port"`
	GRPCPort int    `mapstructure:"grpc_port" yaml:"grpc_port"`
	Host     string `mapstructure:"host" yaml:"host"`
}

// GRPCConfig gRPC 配置
type GRPCConfig struct {
	Server GRPCServerConfig `mapstructure:"server" yaml:"server"`
	Client GRPCClientConfig `mapstructure:"client" yaml:"client"`
}

// GRPCServerConfig gRPC 服务端配置
type GRPCServerConfig struct {
	// 消息大小限制
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize int `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
	
	// 连接配置
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams" yaml:"max_concurrent_streams"`
	ConnectionTimeout    int    `mapstructure:"connection_timeout" yaml:"connection_timeout"`     // 秒
	KeepaliveTime        int    `mapstructure:"keepalive_time" yaml:"keepalive_time"`             // 秒
	KeepaliveTimeout     int    `mapstructure:"keepalive_timeout" yaml:"keepalive_timeout"`       // 秒
	KeepaliveMinTime     int    `mapstructure:"keepalive_min_time" yaml:"keepalive_min_time"`     // 秒
	
	// 安全配置
	EnableReflection bool `mapstructure:"enable_reflection" yaml:"enable_reflection"`
	
	// 压缩配置
	EnableCompression bool   `mapstructure:"enable_compression" yaml:"enable_compression"`
	CompressionLevel  string `mapstructure:"compression_level" yaml:"compression_level"` // gzip, deflate
	
	// 拦截器配置
	EnableLogging  bool `mapstructure:"enable_logging" yaml:"enable_logging"`
	EnableMetrics  bool `mapstructure:"enable_metrics" yaml:"enable_metrics"`
	EnableRecovery bool `mapstructure:"enable_recovery" yaml:"enable_recovery"`
	EnableTracing  bool `mapstructure:"enable_tracing" yaml:"enable_tracing"`
}

// GRPCClientConfig gRPC 客户端配置
type GRPCClientConfig struct {
	// 基础配置
	Timeout        int    `mapstructure:"timeout" yaml:"timeout"`
	MaxRetries     int    `mapstructure:"max_retries" yaml:"max_retries"`
	LoadBalancing  string `mapstructure:"load_balancing" yaml:"load_balancing"`
	
	// 连接配置
	MaxRecvMsgSize       int  `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize       int  `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
	KeepaliveTime        int  `mapstructure:"keepalive_time" yaml:"keepalive_time"`         // 秒
	KeepaliveTimeout     int  `mapstructure:"keepalive_timeout" yaml:"keepalive_timeout"`   // 秒
	PermitWithoutStream  bool `mapstructure:"permit_without_stream" yaml:"permit_without_stream"`
	
	// 重试配置
	RetryPolicy      RetryPolicyConfig `mapstructure:"retry_policy" yaml:"retry_policy"`
	
	// 压缩配置
	EnableCompression bool   `mapstructure:"enable_compression" yaml:"enable_compression"`
	CompressionLevel  string `mapstructure:"compression_level" yaml:"compression_level"`
	
	// 拦截器配置
	EnableLogging bool `mapstructure:"enable_logging" yaml:"enable_logging"`
	EnableMetrics bool `mapstructure:"enable_metrics" yaml:"enable_metrics"`
	EnableTracing bool `mapstructure:"enable_tracing" yaml:"enable_tracing"`
}

// RetryPolicyConfig 重试策略配置
type RetryPolicyConfig struct {
	MaxAttempts          int      `mapstructure:"max_attempts" yaml:"max_attempts"`
	InitialBackoff       string   `mapstructure:"initial_backoff" yaml:"initial_backoff"`       // 如 "1s"
	MaxBackoff           string   `mapstructure:"max_backoff" yaml:"max_backoff"`               // 如 "30s"
	BackoffMultiplier    float64  `mapstructure:"backoff_multiplier" yaml:"backoff_multiplier"`
	RetryableStatusCodes []string `mapstructure:"retryable_status_codes" yaml:"retryable_status_codes"`
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	Type      string   `mapstructure:"type" yaml:"type"`
	Endpoints []string `mapstructure:"endpoints" yaml:"endpoints"`
	Namespace string   `mapstructure:"namespace" yaml:"namespace"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level" yaml:"level"`
	Format string `mapstructure:"format" yaml:"format"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	CertFile string `mapstructure:"cert_file" yaml:"cert_file"`
	KeyFile  string `mapstructure:"key_file" yaml:"key_file"`
	CAFile   string `mapstructure:"ca_file" yaml:"ca_file"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Port    int    `mapstructure:"port" yaml:"port"`
	Path    string `mapstructure:"path" yaml:"path"`
}

var globalConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()
	
	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("application")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}
	
	// 设置环境变量前缀
	v.SetEnvPrefix("GRPC_KIT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// 设置默认值
	setDefaults(v)
	
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	globalConfig = &config
	return &config, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		// 如果没有加载配置，使用默认配置
		config := &Config{}
		setDefaultValues(config)
		globalConfig = config
	}
	return globalConfig
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.host", "0.0.0.0")
	
	// gRPC 服务端默认值
	v.SetDefault("grpc.server.max_recv_msg_size", 4*1024*1024) // 4MB
	v.SetDefault("grpc.server.max_send_msg_size", 4*1024*1024) // 4MB
	v.SetDefault("grpc.server.max_concurrent_streams", 100)
	v.SetDefault("grpc.server.connection_timeout", 120)
	v.SetDefault("grpc.server.keepalive_time", 30)
	v.SetDefault("grpc.server.keepalive_timeout", 5)
	v.SetDefault("grpc.server.keepalive_min_time", 5)
	v.SetDefault("grpc.server.enable_reflection", false)
	v.SetDefault("grpc.server.enable_compression", false)
	v.SetDefault("grpc.server.compression_level", "gzip")
	v.SetDefault("grpc.server.enable_logging", true)
	v.SetDefault("grpc.server.enable_metrics", true)
	v.SetDefault("grpc.server.enable_recovery", true)
	v.SetDefault("grpc.server.enable_tracing", false)
	
	// gRPC 客户端默认值
	v.SetDefault("grpc.client.timeout", 30)
	v.SetDefault("grpc.client.max_retries", 3)
	v.SetDefault("grpc.client.load_balancing", "round_robin")
	v.SetDefault("grpc.client.max_recv_msg_size", 4*1024*1024) // 4MB
	v.SetDefault("grpc.client.max_send_msg_size", 4*1024*1024) // 4MB
	v.SetDefault("grpc.client.keepalive_time", 30)
	v.SetDefault("grpc.client.keepalive_timeout", 5)
	v.SetDefault("grpc.client.permit_without_stream", false)
	v.SetDefault("grpc.client.enable_compression", false)
	v.SetDefault("grpc.client.compression_level", "gzip")
	v.SetDefault("grpc.client.enable_logging", true)
	v.SetDefault("grpc.client.enable_metrics", true)
	v.SetDefault("grpc.client.enable_tracing", false)
	
	// 重试策略默认值
	v.SetDefault("grpc.client.retry_policy.max_attempts", 3)
	v.SetDefault("grpc.client.retry_policy.initial_backoff", "1s")
	v.SetDefault("grpc.client.retry_policy.max_backoff", "30s")
	v.SetDefault("grpc.client.retry_policy.backoff_multiplier", 2.0)
	v.SetDefault("grpc.client.retry_policy.retryable_status_codes", []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"})
	
	v.SetDefault("discovery.type", "etcd")
	v.SetDefault("discovery.endpoints", []string{"localhost:2379"})
	v.SetDefault("discovery.namespace", "/grpc-kit")
	
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	
	v.SetDefault("tls.enabled", false)
	
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 8081)
	v.SetDefault("metrics.path", "/metrics")
	
	v.SetDefault("auto_register.enabled", false)
	v.SetDefault("auto_register.scan_dirs", []string{"./pkg/services", "./internal/services"})
	v.SetDefault("auto_register.patterns", []string{"*.go"})
	v.SetDefault("auto_register.excludes", []string{"*_test.go", "*_mock.go"})
	v.SetDefault("auto_register.service_name", "")
}

// setDefaultValues 设置结构体默认值
func setDefaultValues(config *Config) {
	config.Server.Port = 8080
	config.Server.GRPCPort = 9090
	config.Server.Host = "0.0.0.0"
	
	// gRPC 服务端默认值
	config.GRPC.Server.MaxRecvMsgSize = 4 * 1024 * 1024
	config.GRPC.Server.MaxSendMsgSize = 4 * 1024 * 1024
	config.GRPC.Server.MaxConcurrentStreams = 100
	config.GRPC.Server.ConnectionTimeout = 120
	config.GRPC.Server.KeepaliveTime = 30
	config.GRPC.Server.KeepaliveTimeout = 5
	config.GRPC.Server.KeepaliveMinTime = 5
	config.GRPC.Server.EnableReflection = false
	config.GRPC.Server.EnableCompression = false
	config.GRPC.Server.CompressionLevel = "gzip"
	config.GRPC.Server.EnableLogging = true
	config.GRPC.Server.EnableMetrics = true
	config.GRPC.Server.EnableRecovery = true
	config.GRPC.Server.EnableTracing = false
	
	// gRPC 客户端默认值
	config.GRPC.Client.Timeout = 30
	config.GRPC.Client.MaxRetries = 3
	config.GRPC.Client.LoadBalancing = "round_robin"
	config.GRPC.Client.MaxRecvMsgSize = 4 * 1024 * 1024
	config.GRPC.Client.MaxSendMsgSize = 4 * 1024 * 1024
	config.GRPC.Client.KeepaliveTime = 30
	config.GRPC.Client.KeepaliveTimeout = 5
	config.GRPC.Client.PermitWithoutStream = false
	config.GRPC.Client.EnableCompression = false
	config.GRPC.Client.CompressionLevel = "gzip"
	config.GRPC.Client.EnableLogging = true
	config.GRPC.Client.EnableMetrics = true
	config.GRPC.Client.EnableTracing = false
	
	// 重试策略默认值
	config.GRPC.Client.RetryPolicy.MaxAttempts = 3
	config.GRPC.Client.RetryPolicy.InitialBackoff = "1s"
	config.GRPC.Client.RetryPolicy.MaxBackoff = "30s"
	config.GRPC.Client.RetryPolicy.BackoffMultiplier = 2.0
	config.GRPC.Client.RetryPolicy.RetryableStatusCodes = []string{"UNAVAILABLE", "DEADLINE_EXCEEDED"}
	
	config.Discovery.Type = "etcd"
	config.Discovery.Endpoints = []string{"localhost:2379"}
	config.Discovery.Namespace = "/grpc-kit"
	
	config.Logging.Level = "info"
	config.Logging.Format = "json"
	
	config.TLS.Enabled = false
	
	config.Metrics.Enabled = true
	config.Metrics.Port = 8081
	config.Metrics.Path = "/metrics"
	
	config.AutoRegister.Enabled = false
	config.AutoRegister.ScanDirs = []string{"./pkg/services", "./internal/services"}
	config.AutoRegister.Patterns = []string{"*.go"}
	config.AutoRegister.Excludes = []string{"*_test.go", "*_mock.go"}
	config.AutoRegister.ServiceName = ""
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}