package config

// AutoRegisterConfig 自动注册配置
type AutoRegisterConfig struct {
	Enabled     bool     `mapstructure:"enabled" yaml:"enabled"`           // 是否启用自动注册
	ScanDirs    []string `mapstructure:"scan_dirs" yaml:"scan_dirs"`       // 扫描目录列表
	Patterns    []string `mapstructure:"patterns" yaml:"patterns"`         // 文件匹配模式
	Excludes    []string `mapstructure:"excludes" yaml:"excludes"`         // 排除模式
	ServiceName string   `mapstructure:"service_name" yaml:"service_name"` // 服务名称模式
}