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
