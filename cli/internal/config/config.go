// Package config 管理 CLI 客户端配置
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Config CLI 配置结构
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Device DeviceConfig `mapstructure:"device"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	URL   string `mapstructure:"url"`    // HTTP API 地址
	WSURL string `mapstructure:"ws_url"` // WebSocket 地址
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	AccessToken  string `mapstructure:"access_token"`   // 用户访问 Token（用于 REST）
	RefreshToken string `mapstructure:"refresh_token"`  // 刷新 Token
	DesktopToken string `mapstructure:"desktop_token"`  // 桌面专用 Token（用于 WS）
	Name         string `mapstructure:"name"`           // 设备名称
	ID           string `mapstructure:"id"`             // 设备 ID
}

var (
	cfg        *Config
	configPath string
	configDir  string
)

// Init 初始化配置
func Init() error {
	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	// 配置目录
	configDir = filepath.Join(home, ".pocket-coder")
	configPath = filepath.Join(configDir, "config.yaml")

	// 创建配置目录
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 设置 viper
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	viper.SetDefault("server.url", "http://localhost:8080")
	viper.SetDefault("server.ws_url", "ws://localhost:8080")
	viper.SetDefault("device.access_token", "")
	viper.SetDefault("device.refresh_token", "")
	viper.SetDefault("device.desktop_token", "")
	viper.SetDefault("device.name", getHostname())
	viper.SetDefault("device.id", "")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果文件不存在，创建默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				// 忽略文件已存在的错误
			}
		}
	}

	// 解析配置
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

// Get 获取配置
func Get() *Config {
	return cfg
}

// SaveAuth 保存用户访问/刷新 Token
func SaveAuth(accessToken, refreshToken string) error {
	viper.Set("device.access_token", accessToken)
	viper.Set("device.refresh_token", refreshToken)
	if cfg != nil {
		cfg.Device.AccessToken = accessToken
		cfg.Device.RefreshToken = refreshToken
	}
	return viper.WriteConfig()
}

// SaveDesktop 保存桌面 token 和 ID
func SaveDesktop(desktopToken, desktopID string, name string) error {
	viper.Set("device.desktop_token", desktopToken)
	viper.Set("device.id", desktopID)
	viper.Set("device.name", name)
	if cfg != nil {
		cfg.Device.DesktopToken = desktopToken
		cfg.Device.ID = desktopID
		cfg.Device.Name = name
	}
	return viper.WriteConfig()
}

// GetAccessToken 获取访问 Token
func GetAccessToken() string {
	if cfg == nil {
		return ""
	}
	return cfg.Device.AccessToken
}

// GetDesktopToken 获取桌面 Token（用于 WS）
func GetDesktopToken() string {
	if cfg == nil {
		return ""
	}
	return cfg.Device.DesktopToken
}

// GetDesktopID 获取设备 ID
func GetDesktopID() string {
	if cfg == nil {
		return ""
	}
	return cfg.Device.ID
}

// GetServerURL 获取服务器地址
func GetServerURL() string {
	if cfg == nil {
		return "http://localhost:8080"
	}
	return cfg.Server.URL
}

// ClearToken 清除本地凭证
func ClearToken() error {
	viper.Set("device.access_token", "")
	viper.Set("device.refresh_token", "")
	viper.Set("device.desktop_token", "")
	viper.Set("device.id", "")
	if cfg != nil {
		cfg.Device.AccessToken = ""
		cfg.Device.RefreshToken = ""
		cfg.Device.DesktopToken = ""
		cfg.Device.ID = ""
	}
	return viper.WriteConfig()
}

// SetServerURL 设置服务器地址
func SetServerURL(url string) {
	viper.Set("server.url", url)
	// 自动设置 WebSocket 地址
	wsURL := "ws" + url[4:] // http -> ws, https -> wss
	viper.Set("server.ws_url", wsURL)
	if cfg != nil {
		cfg.Server.URL = url
		cfg.Server.WSURL = wsURL
	}
}

// IsLoggedIn 检查是否已登录
func IsLoggedIn() bool {
	return cfg != nil && cfg.Device.AccessToken != ""
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// GetDeviceUUID 获取或生成设备唯一标识
// 该 UUID 持久化存储在 ~/.pocket-coder/device_id 文件中
// 即使用户更改主机名，设备 UUID 也不会变化
func GetDeviceUUID() (string, error) {
	deviceIDPath := filepath.Join(configDir, "device_id")

	// 尝试读取现有的 device_id
	data, err := os.ReadFile(deviceIDPath)
	if err == nil {
		deviceUUID := string(data)
		if deviceUUID != "" {
			return deviceUUID, nil
		}
	}

	// 如果不存在或为空，生成新的 UUID
	newUUID := uuid.New().String()

	// 持久化保存
	if err := os.WriteFile(deviceIDPath, []byte(newUUID), 0600); err != nil {
		return "", fmt.Errorf("保存设备 UUID 失败: %w", err)
	}

	return newUUID, nil
}
