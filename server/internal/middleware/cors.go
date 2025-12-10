// Package middleware 提供 HTTP 请求的中间件
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 跨域配置
type CORSConfig struct {
	AllowOrigins     []string // 允许的来源，如 ["http://localhost:3000", "https://example.com"]
	AllowMethods     []string // 允许的 HTTP 方法
	AllowHeaders     []string // 允许的请求头
	ExposeHeaders    []string // 允许暴露的响应头
	AllowCredentials bool     // 是否允许携带凭据（Cookie）
	MaxAge           int      // 预检请求结果的缓存时间（秒）
}

// DefaultCORSConfig 返回默认的 CORS 配置
// 默认允许所有来源，适用于开发环境
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"}, // 允许所有来源（生产环境应该限制）
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 小时
	}
}

// CORSMiddleware 创建 CORS 跨域中间件
// 参数:
//   - config: CORS 配置，传入空值使用默认配置
//
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
func CORSMiddleware(config ...CORSConfig) gin.HandlerFunc {
	// 如果没有传入配置，使用默认配置
	var cfg CORSConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查来源是否被允许
		allowOrigin := ""
		if len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*" {
			// 允许所有来源
			allowOrigin = "*"
		} else {
			// 检查请求来源是否在允许列表中
			for _, o := range cfg.AllowOrigins {
				if o == origin {
					allowOrigin = origin
					break
				}
			}
		}

		// 如果来源被允许，设置 CORS 响应头
		if allowOrigin != "" {
			// Access-Control-Allow-Origin: 允许的来源
			c.Header("Access-Control-Allow-Origin", allowOrigin)

			// Access-Control-Allow-Credentials: 是否允许携带凭据
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// Access-Control-Expose-Headers: 允许浏览器访问的响应头
			if len(cfg.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders))
			}
		}

		// 处理预检请求（OPTIONS）
		// 浏览器在发送"非简单请求"前，会先发送 OPTIONS 请求检查服务器是否允许
		if c.Request.Method == http.MethodOptions {
			// Access-Control-Allow-Methods: 允许的 HTTP 方法
			c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods))

			// Access-Control-Allow-Headers: 允许的请求头
			c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders))

			// Access-Control-Max-Age: 预检请求结果的缓存时间
			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", itoa(cfg.MaxAge))
			}

			// 预检请求直接返回 204，不继续处理
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 继续处理实际请求
		c.Next()
	}
}

// joinStrings 将字符串切片用逗号连接
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

// itoa 将整数转换为字符串（简单实现）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}

	if negative {
		result = "-" + result
	}
	return result
}
