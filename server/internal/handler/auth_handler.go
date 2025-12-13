// Package handler 提供 HTTP 请求处理器
// 对应 Java 中的 Controller 层
package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/pkg/response"
)

// AuthHandler 认证请求处理器
// 处理用户注册、登录、登出以及设备授权
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body service.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=service.RegisterResponse}
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// 1. 解析请求参数
	var req service.RegisterRequest
	// ShouldBindJSON 会自动验证 binding 标签中的规则
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用服务层处理注册
	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		// 根据错误类型返回不同的响应
		switch err {
		case service.ErrInvalidUsername:
			response.BadRequest(c, "用户名只能包含字母、数字和下划线，长度3-20")
		case service.ErrUserExists:
			response.UserExists(c)
		case service.ErrEmailExists:
			response.ErrorWithCode(c, 400, response.CodeBadRequest, "邮箱已被注册")
		case service.ErrPhoneExists:
			response.ErrorWithCode(c, 400, response.CodeBadRequest, "手机号已被注册")
		default:
			response.InternalError(c, "注册失败")
		}
		return
	}

	// 3. 返回成功响应
	response.SuccessWithMessage(c, "注册成功", result)
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body service.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.UserNotFound(c)
		case service.ErrPasswordWrong:
			response.PasswordWrong(c)
		default:
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.SuccessWithMessage(c, "登录成功", result)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 登出当前用户，将 Token 加入黑名单
// @Tags 认证
// @Security Bearer
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取 Token 信息（由认证中间件设置）
	token, exists := c.Get("token")
	if !exists {
		response.BadRequest(c, "无法获取 Token 信息")
		return
	}

	expireAt, exists := c.Get("token_exp")
	if !exists {
		response.BadRequest(c, "无法获取 Token 过期时间")
		return
	}

	// 计算 Token 哈希
	tokenHash := hashToken(token.(string))

	// 将 Token 加入黑名单
	// expireAt 是 *jwt.NumericDate 类型
	if err := h.authService.Logout(c.Request.Context(), tokenHash, expireAt.(time.Time)); err != nil {
		response.InternalError(c, "登出失败")
		return
	}

	response.SuccessWithMessage(c, "登出成功", nil)
}

// RefreshToken 刷新 Token
// @Summary 刷新 Token
// @Description 使用 Refresh Token 获取新的 Access Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} response.Response{data=service.RefreshTokenResponse}
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Refresh Token 无效或已过期")
		return
	}

	response.Success(c, result)
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// hashToken 计算 Token 哈希
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
