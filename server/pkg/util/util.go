// Package util 提供通用工具函数
package util

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 哈希密码
// bcrypt 是一种专门为密码哈希设计的算法，自动添加盐值
// 参数:
//   - password: 明文密码
//
// 返回:
//   - string: 密码哈希值
//   - error: 哈希错误
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost 是默认的计算成本（10）
	// 成本越高，计算越慢，安全性越高
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码是否匹配
// 参数:
//   - password: 用户输入的明文密码
//   - hash: 数据库中存储的哈希值
//
// 返回:
//   - bool: 是否匹配
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateUUID 生成 UUID
// 使用 Google 的 uuid 库生成 UUID v4
// 返回:
//   - string: UUID 字符串（不含连字符）
func GenerateUUID() string {
	// uuid.New() 生成 UUID v4（随机生成）
	// String() 返回格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// 我们去掉连字符使其更紧凑
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenerateDeviceToken 生成设备令牌
// 64 字符的随机十六进制字符串
// 返回:
//   - string: 设备令牌
func GenerateDeviceToken() string {
	// 32 字节 = 64 个十六进制字符
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateDeviceCode 生成设备授权码（长码）
// 用于电脑端内部使用
// 返回:
//   - string: 设备码
func GenerateDeviceCode() string {
	// 20 字节 = 40 个十六进制字符
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateUserCode 生成用户授权码（短码）
// 用于用户手动输入，格式：XXXX-XXXX
// 只使用易于区分的大写字母和数字
// 返回:
//   - string: 用户码，如 "ABCD-1234"
func GenerateUserCode() string {
	// 使用易于区分的字符，避免 0/O、1/I/L 等容易混淆的字符
	const chars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	
	result := make([]byte, 8)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	
	// 在中间插入连字符，更易于阅读
	return string(result[:4]) + "-" + string(result[4:])
}

// GenerateRandomString 生成指定长度的随机字符串
// 参数:
//   - length: 字符串长度
//
// 返回:
//   - string: 随机字符串
func GenerateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// TruncateString 截断字符串到指定长度
// 如果字符串超过指定长度，截断并添加 "..."
// 参数:
//   - s: 原字符串
//   - maxLen: 最大长度
//
// 返回:
//   - string: 截断后的字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// StringPtr 返回字符串的指针
// 用于可选字段的赋值
// 参数:
//   - s: 字符串
//
// 返回:
//   - *string: 字符串指针
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr 返回 int64 的指针
func Int64Ptr(i int64) *int64 {
	return &i
}

// IntPtr 返回 int 的指针
func IntPtr(i int) *int {
	return &i
}

// BoolPtr 返回 bool 的指针
func BoolPtr(b bool) *bool {
	return &b
}
