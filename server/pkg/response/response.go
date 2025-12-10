// Package response 提供统一的 HTTP 响应格式
// 所有 API 都使用相同的响应结构，便于前端处理
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
// code: 业务状态码（0 表示成功）
// message: 提示信息
// data: 响应数据
type Response struct {
	Code    int         `json:"code"`              // 业务状态码
	Message string      `json:"message"`           // 提示信息
	Data    interface{} `json:"data,omitempty"`    // 响应数据，可选
}

// 业务状态码定义
const (
	CodeSuccess          = 0    // 成功
	CodeBadRequest       = 1000 // 请求参数错误
	CodeUnauthorized     = 1001 // 未授权
	CodeForbidden        = 1002 // 禁止访问
	CodeNotFound         = 1003 // 资源不存在
	CodeInternalError    = 1004 // 服务器内部错误
	CodeUserExists       = 1101 // 用户已存在
	CodeUserNotFound     = 1102 // 用户不存在
	CodePasswordWrong    = 1103 // 密码错误
	CodeDeviceNotFound   = 1201 // 设备不存在
	CodeDeviceOffline    = 1202 // 设备离线
	CodeSessionNotFound  = 1301 // 会话不存在
	CodeSessionEnded     = 1302 // 会话已结束
	CodeDeviceCodeExpired = 1401 // 设备授权码过期
)

// Success 返回成功响应
// 参数:
//   - c: Gin 上下文
//   - data: 响应数据，可以是任意类型
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 返回成功响应（带自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 返回错误响应
// 参数:
//   - c: Gin 上下文
//   - httpCode: HTTP 状态码
//   - message: 错误信息
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// ErrorWithCode 返回错误响应（带业务状态码）
// 参数:
//   - c: Gin 上下文
//   - httpCode: HTTP 状态码
//   - bizCode: 业务状态码
//   - message: 错误信息
func ErrorWithCode(c *gin.Context, httpCode, bizCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    bizCode,
		Message: message,
	})
}

// BadRequest 返回 400 错误（请求参数错误）
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
	})
}

// Unauthorized 返回 401 错误（未授权）
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// Fail 返回失败响应（通用）
// 参数:
//   - c: Gin 上下文
//   - httpCode: HTTP 状态码
//   - message: 错误信息
func Fail(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// Forbidden 返回 403 错误（禁止访问）
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound 返回 404 错误（资源不存在）
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// InternalError 返回 500 错误（服务器内部错误）
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
	})
}

// UserExists 返回用户已存在错误
func UserExists(c *gin.Context) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeUserExists,
		Message: "用户名已存在",
	})
}

// UserNotFound 返回用户不存在错误
func UserNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeUserNotFound,
		Message: "用户不存在",
	})
}

// PasswordWrong 返回密码错误
func PasswordWrong(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodePasswordWrong,
		Message: "密码错误",
	})
}

// DeviceNotFound 返回设备不存在错误
func DeviceNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeDeviceNotFound,
		Message: "设备不存在",
	})
}

// DeviceOffline 返回设备离线错误
func DeviceOffline(c *gin.Context) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeDeviceOffline,
		Message: "设备已离线",
	})
}

// SessionNotFound 返回会话不存在错误
func SessionNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeSessionNotFound,
		Message: "会话不存在",
	})
}

// DeviceCodeExpired 返回设备授权码过期错误
func DeviceCodeExpired(c *gin.Context) {
	c.JSON(http.StatusGone, Response{
		Code:    CodeDeviceCodeExpired,
		Message: "授权码已过期",
	})
}

// Created 返回 201 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeSuccess,
		Message: "创建成功",
		Data:    data,
	})
}

// NoContent 返回 204 无内容响应（用于删除操作）
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Accepted 返回 202 已接受响应（用于异步操作）
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Code:    CodeSuccess,
		Message: "请求已接受",
		Data:    data,
	})
}
