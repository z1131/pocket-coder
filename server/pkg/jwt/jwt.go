// Package jwt 提供 JWT Token 的生成和验证功能
// JWT (JSON Web Token) 用于用户认证和设备认证
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义错误类型
var (
	ErrInvalidToken = errors.New("invalid token")    // Token 无效
	ErrExpiredToken = errors.New("token has expired") // Token 已过期
)

// UserClaims 用户 JWT 的声明（Payload）
// 包含用户相关信息
type UserClaims struct {
	UserID   int64  `json:"user_id"`  // 用户 ID
	Username string `json:"username"` // 用户名
	jwt.RegisteredClaims               // 标准声明（过期时间等）
}

// DesktopClaims 设备 JWT 的声明
// 用于电脑端认证
type DesktopClaims struct {
	UserID      int64  `json:"user_id"`      // 用户 ID
	DesktopID   int64  `json:"desktop_id"`   // 设备 ID
	DeviceToken string `json:"device_token"` // 设备令牌
	jwt.RegisteredClaims
}

// JWTService 提供 JWT 相关操作
type JWTService struct {
	secret        []byte        // JWT 签名密钥
	accessExpire  time.Duration // Access Token 过期时间
	refreshExpire time.Duration // Refresh Token 过期时间
}

// NewJWTService 创建 JWTService 实例
// 参数:
//   - secret: JWT 签名密钥，至少 32 个字符
//   - accessExpire: Access Token 过期时间
//   - refreshExpire: Refresh Token 过期时间
//
// 返回:
//   - *JWTService: JWT 服务实例
func NewJWTService(secret string, accessExpire, refreshExpire time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}
}

// GenerateAccessToken 生成 Access Token
// 用于普通请求的认证
// 参数:
//   - userID: 用户 ID
//   - username: 用户名
//
// 返回:
//   - string: JWT Token 字符串
//   - error: 生成错误
func (s *JWTService) GenerateAccessToken(userID int64, username string) (string, error) {
	// 创建声明
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: Token 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpire)),
			// IssuedAt: Token 签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// NotBefore: Token 生效时间（设为现在）
			NotBefore: jwt.NewNumericDate(time.Now()),
			// Issuer: 签发者标识
			Issuer: "pocket-coder",
			// Subject: 主题（这里使用 "access" 区分 Token 类型）
			Subject: "access",
		},
	}

	// 创建 Token
	// jwt.SigningMethodHS256: 使用 HMAC SHA256 算法签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并返回 Token 字符串
	return token.SignedString(s.secret)
}

// GenerateRefreshToken 生成 Refresh Token
// 用于刷新 Access Token
// 参数:
//   - userID: 用户 ID
//   - username: 用户名
//
// 返回:
//   - string: JWT Token 字符串
//   - error: 生成错误
func (s *JWTService) GenerateRefreshToken(userID int64, username string) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pocket-coder",
			Subject:   "refresh", // 标识为 Refresh Token
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateDesktopToken 生成设备 Token
// 用于电脑端认证
// 参数:
//   - userID: 用户 ID
//   - desktopID: 设备 ID
//   - deviceToken: 设备令牌
//
// 返回:
//   - string: JWT Token 字符串
//   - error: 生成错误
func (s *JWTService) GenerateDesktopToken(userID, desktopID int64, deviceToken string) (string, error) {
	claims := DesktopClaims{
		UserID:      userID,
		DesktopID:   desktopID,
		DeviceToken: deviceToken,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设备 Token 使用较长的过期时间（30 天）
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pocket-coder",
			Subject:   "desktop",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken 验证用户 Token
// 参数:
//   - tokenString: JWT Token 字符串
//
// 返回:
//   - *UserClaims: Token 中的声明信息
//   - error: 验证错误（无效或已过期）
func (s *JWTService) ValidateToken(tokenString string) (*UserClaims, error) {
	// 解析 Token
	// 第二个参数是一个空的 UserClaims 实例，用于接收解析结果
	// 第三个参数是密钥提供函数
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		// 确保使用的是我们期望的算法（HMAC）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		// 检查是否是过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// 类型断言获取 claims
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateDesktopToken 验证设备 Token
// 参数:
//   - tokenString: JWT Token 字符串
//
// 返回:
//   - *DesktopClaims: Token 中的声明信息
//   - error: 验证错误
func (s *JWTService) ValidateDesktopToken(tokenString string) (*DesktopClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &DesktopClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*DesktopClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken 验证 Refresh Token
// 参数:
//   - tokenString: JWT Token 字符串
//
// 返回:
//   - *UserClaims: Token 中的声明信息
//   - error: 验证错误
func (s *JWTService) ValidateRefreshToken(tokenString string) (*UserClaims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 检查是否是 Refresh Token
	if claims.Subject != "refresh" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessExpire 获取 Access Token 过期时间
func (s *JWTService) GetAccessExpire() time.Duration {
	return s.accessExpire
}

// GetRefreshExpire 获取 Refresh Token 过期时间
func (s *JWTService) GetRefreshExpire() time.Duration {
	return s.refreshExpire
}

// ParseUserToken 解析用户 Token（独立函数，供 WebSocket 使用）
// 参数:
//   - tokenString: JWT Token 字符串
//   - secret: JWT 签名密钥
//
// 返回:
//   - *UserClaims: Token 中的声明信息
//   - error: 验证错误
func ParseUserToken(tokenString, secret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ParseDeviceToken 解析设备 Token（独立函数，供 WebSocket 使用）
// 参数:
//   - tokenString: JWT Token 字符串
//   - secret: JWT 签名密钥
//
// 返回:
//   - *DesktopClaims: Token 中的声明信息
//   - error: 验证错误
func ParseDeviceToken(tokenString, secret string) (*DesktopClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &DesktopClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*DesktopClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
