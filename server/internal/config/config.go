// Package config 负责加载和管理应用程序的配置
// 使用 viper 库支持 YAML 配置文件和环境变量覆盖
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用程序的根配置结构
// 包含所有子配置模块
type Config struct {
	Server ServerConfig `mapstructure:"server"` // 服务器配置
	MySQL  MySQLConfig  `mapstructure:"mysql"`  // MySQL 配置
	Redis  RedisConfig  `mapstructure:"redis"`  // Redis 配置
	JWT    JWTConfig    `mapstructure:"jwt"`    // JWT 配置
	Log    LogConfig    `mapstructure:"log"`    // 日志配置
	AI     AIConfig     `mapstructure:"ai"`     // AI 服务配置
}

// AIConfig AI 服务配置
type AIConfig struct {
	QwenAPIKey string `mapstructure:"qwen_api_key"` // Qwen API Key
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Port int      `mapstructure:"port"` // 监听端口，默认 8080
	Mode string   `mapstructure:"mode"` // 运行模式: debug / release
	CORS []string `mapstructure:"cors"` // CORS 允许的域名
}

// MySQLConfig MySQL 数据库连接配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`           // 数据库主机地址
	Port         int    `mapstructure:"port"`           // 数据库端口
	Username     string `mapstructure:"username"`       // 数据库用户名
	Password     string `mapstructure:"password"`       // 数据库密码
	Database     string `mapstructure:"database"`       // 数据库名称
	Charset      string `mapstructure:"charset"`        // 字符集
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 最大打开连接数
	MaxLifetime  int    `mapstructure:"max_lifetime"`   // 连接最大生命周期（秒）
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`      // Redis 主机地址
	Port     int    `mapstructure:"port"`      // Redis 端口
	Username string `mapstructure:"username"`  // Redis 用户名（阿里云需要）
	Password string `mapstructure:"password"`  // Redis 密码
	DB       int    `mapstructure:"db"`        // 数据库索引 (0-15)
	PoolSize int    `mapstructure:"pool_size"` // 连接池大小
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`         // JWT 签名密钥，至少32字符
	AccessExpire  time.Duration `mapstructure:"access_expire"`  // Access Token 过期时间
	RefreshExpire time.Duration `mapstructure:"refresh_expire"` // Refresh Token 过期时间
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别: debug/info/warn/error
	Format string `mapstructure:"format"` // 日志格式: json/text
}

// Load 从指定路径加载配置文件
// 支持环境变量覆盖配置项
// 参数:
//   - configPath: 配置文件目录路径 (如 "./configs")
//
// 返回:
//   - *Config: 配置对象
//   - error: 如果加载失败则返回错误
func Load(configPath string) (*Config, error) {
	// 创建新的 viper 实例
	v := viper.New()

	// 设置配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// 启用环境变量
	v.AutomaticEnv()
	// 将环境变量中的 _ 映射到配置的 .
	// 例如: MYSQL_HOST -> mysql.host
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量
	bindEnvVariables(v)

	// 设置默认值（当配置文件中未指定时使用）
	setDefaults(v)

	// 读取配置文件（如果不存在则使用默认值和环境变量）
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，继续使用默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 将配置解析到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// bindEnvVariables 绑定环境变量到配置项
func bindEnvVariables(v *viper.Viper) {
	// 服务器配置
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")

	// MySQL 配置
	v.BindEnv("mysql.host", "MYSQL_HOST")
	v.BindEnv("mysql.port", "MYSQL_PORT")
	v.BindEnv("mysql.username", "MYSQL_USERNAME")
	v.BindEnv("mysql.password", "MYSQL_PASSWORD")
	v.BindEnv("mysql.database", "MYSQL_DATABASE")

	// Redis 配置
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.username", "REDIS_USERNAME")
	v.BindEnv("redis.password", "REDIS_PASSWORD")

	// JWT 配置
	v.BindEnv("jwt.secret", "JWT_SECRET")

	// AI 配置
	v.BindEnv("ai.qwen_api_key", "QWEN_API_KEY")
}

// setDefaults 设置配置项的默认值
// 当配置文件中没有指定某个值时，将使用这里设置的默认值
func setDefaults(v *viper.Viper) {
	// 服务器默认配置
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.cors", []string{"http://localhost:3000", "http://localhost:5173"})

	// MySQL 默认配置
	v.SetDefault("mysql.host", "localhost")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 100)
	v.SetDefault("mysql.max_lifetime", 3600)

	// Redis 默认配置
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)

	// JWT 默认配置
	v.SetDefault("jwt.access_expire", "24h")
	v.SetDefault("jwt.refresh_expire", "168h")

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}
