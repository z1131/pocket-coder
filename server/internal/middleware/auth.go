// Package middleware 提供 HTTP 请求的中间件
// 包括 JWT 认证、CORS 跨域、日志记录等
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/cache"
	"pocket-coder-server/pkg/jwt"
	"pocket-coder-server/pkg/response"
)

// AuthMiddleware 创建 JWT 认证中间件
// 验证请求头中的 Bearer Token，并将用户信息存入上下文
// 参数:
//   - jwtService: JWT 服务实例，用于解析和验证 Token
//   - redisCache: Redis 缓存实例，用于检查 Token 黑名单
//
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
func AuthMiddleware(jwtService *jwt.JWTService, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取 Authorization 字段
		// 格式: "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有提供认证信息
			response.Unauthorized(c, "请先登录")
			c.Abort() // 终止请求处理
			return
		}

		// 2. 解析 Bearer Token
		// 检查是否以 "Bearer " 开头
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 格式不正确
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 3. 验证 Token
		// 解析 JWT 并验证签名和过期时间
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			// Token 无效或已过期
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 4. 检查 Token 是否在黑名单中
		// 用户登出后，Token 会被加入黑名单
		// 计算 Token 的哈希值（不存储原始 Token）
		tokenHash := hashToken(tokenString)
		if redisCache.IsTokenBlacklisted(c.Request.Context(), tokenHash) {
			// Token 已被加入黑名单（用户已登出）
			response.Unauthorized(c, "Token 已失效，请重新登录")
			c.Abort()
			return
		}

		// 5. 将用户信息存入上下文
		// 后续的 Handler 可以通过 c.GetInt64("user_id") 获取
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", tokenString)           // 存储原始 Token，用于登出时计算哈希
		c.Set("token_exp", claims.ExpiresAt)  // 存储过期时间，用于登出时设置黑名单 TTL

		// 6. 继续处理请求
		c.Next()
	}
}

// OptionalAuthMiddleware 创建可选的 JWT 认证中间件
// 与 AuthMiddleware 类似，但不强制要求认证
// 如果提供了有效 Token，会将用户信息存入上下文
// 如果没有提供或 Token 无效，仍然继续处理请求
func OptionalAuthMiddleware(jwtService *jwt.JWTService, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		tokenString := parts[1]

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		tokenHash := hashToken(tokenString)
		if redisCache.IsTokenBlacklisted(c.Request.Context(), tokenHash) {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", tokenString)
		c.Set("token_exp", claims.ExpiresAt)

		c.Next()
	}
}

// DesktopAuthMiddleware 创建设备认证中间件
// 用于验证电脑端的请求
// 参数:
//   - jwtService: JWT 服务实例
//   - redisCache: Redis 缓存实例
//
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
func DesktopAuthMiddleware(jwtService *jwt.JWTService, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录设备")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 验证设备 Token
		claims, err := jwtService.ValidateDesktopToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "设备 Token 无效或已过期")
			c.Abort()
			return
		}

		// 检查黑名单
		tokenHash := hashToken(tokenString)
		if redisCache.IsTokenBlacklisted(c.Request.Context(), tokenHash) {
			response.Unauthorized(c, "设备 Token 已失效")
			c.Abort()
			return
		}

		// 将设备信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("desktop_id", claims.DesktopID)
		c.Set("device_token", claims.DeviceToken)
		c.Set("token", tokenString)
		c.Set("token_exp", claims.ExpiresAt)

		c.Next()
	}
}

// hashToken 计算 Token 的 SHA256 哈希值
// 用于黑名单存储，避免存储原始 Token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetUserID 从上下文获取用户 ID 的辅助函数
// 参数:
//   - c: Gin 上下文
//
// 返回:
//   - int64: 用户 ID，如果未认证返回 0
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int64)
}

// GetDesktopID 从上下文获取设备 ID 的辅助函数
// 参数:
//   - c: Gin 上下文
//
// 返回:
//   - int64: 设备 ID，如果不是设备认证返回 0
func GetDesktopID(c *gin.Context) int64 {
	desktopID, exists := c.Get("desktop_id")
	if !exists {
		return 0
	}
	return desktopID.(int64)
}

// GetUsername 从上下文获取用户名的辅助函数
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// RequireAuth 检查用户是否已认证的辅助函数
// 如果未认证，返回 401 错误并终止请求
func RequireAuth(c *gin.Context) bool {
	userID := GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "请先登录")
		c.Abort()
		return false
	}
	return true
}
