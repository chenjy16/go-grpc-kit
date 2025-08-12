package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server    ServerConfig    `mapstructure:"server" yaml:"server"`
	GRPC      GRPCConfig      `mapstructure:"grpc" yaml:"grpc"`
	Discovery DiscoveryConfig `mapstructure:"discovery" yaml:"discovery"`
	Logging   LoggingConfig   `mapstructure:"logging" yaml:"logging"`
	TLS       TLSConfig       `mapstructure:"tls" yaml:"tls"`
	Metrics   MetricsConfig   `mapstructure:"metrics" yaml:"metrics"`
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
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize int `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
}

// GRPCClientConfig gRPC 客户端配置
type GRPCClientConfig struct {
	Timeout        int    `mapstructure:"timeout" yaml:"timeout"`
	MaxRetries     int    `mapstructure:"max_retries" yaml:"max_retries"`
	LoadBalancing  string `mapstructure:"load_balancing" yaml:"load_balancing"`
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
	
	v.SetDefault("grpc.server.max_recv_msg_size", 4*1024*1024) // 4MB
	v.SetDefault("grpc.server.max_send_msg_size", 4*1024*1024) // 4MB
	
	v.SetDefault("grpc.client.timeout", 30)
	v.SetDefault("grpc.client.max_retries", 3)
	v.SetDefault("grpc.client.load_balancing", "round_robin")
	
	v.SetDefault("discovery.type", "etcd")
	v.SetDefault("discovery.endpoints", []string{"localhost:2379"})
	v.SetDefault("discovery.namespace", "/grpc-kit")
	
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	
	v.SetDefault("tls.enabled", false)
	
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 8081)
	v.SetDefault("metrics.path", "/metrics")
}

// setDefaultValues 设置结构体默认值
func setDefaultValues(config *Config) {
	config.Server.Port = 8080
	config.Server.GRPCPort = 9090
	config.Server.Host = "0.0.0.0"
	
	config.GRPC.Server.MaxRecvMsgSize = 4 * 1024 * 1024
	config.GRPC.Server.MaxSendMsgSize = 4 * 1024 * 1024
	
	config.GRPC.Client.Timeout = 30
	config.GRPC.Client.MaxRetries = 3
	config.GRPC.Client.LoadBalancing = "round_robin"
	
	config.Discovery.Type = "etcd"
	config.Discovery.Endpoints = []string{"localhost:2379"}
	config.Discovery.Namespace = "/grpc-kit"
	
	config.Logging.Level = "info"
	config.Logging.Format = "json"
	
	config.TLS.Enabled = false
	
	config.Metrics.Enabled = true
	config.Metrics.Port = 8081
	config.Metrics.Path = "/metrics"
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}