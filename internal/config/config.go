package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 存储了应用程序的所有配置。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	FRP      FRPConfig      `mapstructure:"frp"`
}

// ServerConfig 存储了 API 服务器的配置。
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Addr string `mapstructure:"addr"`
}

// DatabaseConfig 存储了数据库连接的配置。
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// JWTConfig 存储了 JWT 认证的配置。
type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	TokenTTL  int    `mapstructure:"token_ttl"`
}

// FRPConfig 存储了 FRPS 隧道的配置。
type FRPConfig struct {
	BindPort      int    `mapstructure:"bind_port"`
	Token         string `mapstructure:"token"`
	DashboardPort int    `mapstructure:"dashboard_port"`
	DashboardUser string `mapstructure:"dashboard_user"`
	DashboardPwd  string `mapstructure:"dashboard_pwd"`
	DashboardAddr string `mapstructure:"dashboard_addr"`
	AgentToken    string `mapstructure:"agent_token"`
}

// Load 从文件和环境变量中加载配置。
func Load() (*Config, error) {
	v := viper.New()

	// 设置默认值
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.addr", "0.0.0.0")
	v.SetDefault("jwt.token_ttl", 3600) // 1 hour
	v.SetDefault("frp.bind_port", 7000)

	// 设置配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")     // For running from root
	v.AddConfigPath("../configs")    // For running from cmd/server
	v.AddConfigPath("../../configs") // For running from internal/api, etc.
	v.AddConfigPath(".")

	// 从环境变量中读取
	v.AutomaticEnv()
	v.SetEnvPrefix("UTOPIA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 不再区分错误类型，任何读取错误都应被报告
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
