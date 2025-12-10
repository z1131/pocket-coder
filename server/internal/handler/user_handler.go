// Package handler 提供 HTTP 请求处理器
package handler

import (
	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/pkg/response"
)

// UserHandler 用户请求处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取用户资料
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Security Bearer
// @Produce json
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户 ID（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID.(int64))
	if err != nil {
		if err == service.ErrUserNotFound {
			response.UserNotFound(c)
			return
		}
		response.InternalError(c, "获取用户信息失败")
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body service.UpdateProfileRequest true "要更新的字段"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.UserNotFound(c)
		case service.ErrEmailExists:
			response.ErrorWithCode(c, 400, response.CodeBadRequest, "邮箱已被其他用户使用")
		default:
			response.InternalError(c, "更新用户信息失败")
		}
		return
	}

	response.Success(c, user)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body service.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.UserNotFound(c)
		case service.ErrPasswordWrong:
			response.ErrorWithCode(c, 400, response.CodePasswordWrong, "旧密码错误")
		default:
			response.InternalError(c, "修改密码失败")
		}
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}
