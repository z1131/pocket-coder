// Package middleware 提供 HTTP 请求的中间件
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 创建请求日志中间件
// 记录每个请求的方法、路径、状态码和耗时
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 获取请求路径
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 获取错误信息（如果有）
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 根据状态码选择日志级别
		// 200-299: 成功
		// 300-399: 重定向
		// 400-499: 客户端错误
		// 500-599: 服务端错误
		logLine := formatLogLine(statusCode, latency, clientIP, method, path, errorMessage)

		if statusCode >= 500 {
			// 服务端错误，使用错误级别日志
			log.Printf("[ERROR] %s", logLine)
		} else if statusCode >= 400 {
			// 客户端错误，使用警告级别日志
			log.Printf("[WARN] %s", logLine)
		} else {
			// 正常请求，使用信息级别日志
			log.Printf("[INFO] %s", logLine)
		}
	}
}

// formatLogLine 格式化日志行
func formatLogLine(statusCode int, latency time.Duration, clientIP, method, path, errorMessage string) string {
	// 格式化耗时
	// 小于 1ms 显示微秒
	// 小于 1s 显示毫秒
	// 否则显示秒
	var latencyStr string
	if latency < time.Millisecond {
		latencyStr = latency.String()
	} else if latency < time.Second {
		latencyStr = latency.Truncate(time.Microsecond).String()
	} else {
		latencyStr = latency.Truncate(time.Millisecond).String()
	}

	// 基本日志格式
	logLine := statusCodeColor(statusCode) + " | " +
		padRight(latencyStr, 12) + " | " +
		padRight(clientIP, 15) + " | " +
		padRight(method, 7) + " | " +
		path

	// 如果有错误信息，追加到日志
	if errorMessage != "" {
		logLine += " | " + errorMessage
	}

	return logLine
}

// statusCodeColor 根据状态码返回带颜色标记的状态码
// 注意：这里简化处理，不使用 ANSI 颜色码
func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "[" + itoa(code) + " OK]"
	case code >= 300 && code < 400:
		return "[" + itoa(code) + " REDIRECT]"
	case code >= 400 && code < 500:
		return "[" + itoa(code) + " CLIENT_ERR]"
	default:
		return "[" + itoa(code) + " SERVER_ERR]"
	}
}

// padRight 右填充字符串到指定长度
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := length - len(s)
	for i := 0; i < padding; i++ {
		s += " "
	}
	return s
}

// RecoveryMiddleware 创建 panic 恢复中间件
// 捕获处理器中的 panic，防止程序崩溃
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 信息
				log.Printf("[PANIC] %v", err)

				// 返回 500 错误
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": "服务器内部错误",
				})
			}
		}()

		c.Next()
	}
}
