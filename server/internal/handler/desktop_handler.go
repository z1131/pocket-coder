// Package handler 提供 HTTP 请求处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/pkg/jwt"
	"pocket-coder-server/pkg/response"
)

// DesktopHandler 设备请求处理器
type DesktopHandler struct {
	desktopService *service.DesktopService
	jwtService     *jwt.JWTService
}

// NewDesktopHandler 创建 DesktopHandler 实例
func NewDesktopHandler(desktopService *service.DesktopService, jwtService *jwt.JWTService) *DesktopHandler {
	return &DesktopHandler{
		desktopService: desktopService,
		jwtService:     jwtService,
	}
}

// ListDesktops 获取设备列表
// @Summary 获取我的设备列表
// @Description 获取当前用户绑定的所有电脑设备
// @Tags 设备
// @Security Bearer
// @Produce json
// @Success 200 {object} response.Response{data=[]service.DesktopResponse}
// @Router /api/desktops [get]
func (h *DesktopHandler) ListDesktops(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	desktops, err := h.desktopService.ListDesktops(c.Request.Context(), userID.(int64))
	if err != nil {
		response.InternalError(c, "获取设备列表失败")
		return
	}

	response.Success(c, gin.H{
		"desktops": desktops,
	})
}

// GetDesktop 获取设备详情
// @Summary 获取设备详情
// @Description 获取指定设备的详细信息
// @Tags 设备
// @Security Bearer
// @Produce json
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response{data=service.DesktopResponse}
// @Router /api/desktops/{id} [get]
func (h *DesktopHandler) GetDesktop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 解析路径参数
	desktopID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的设备ID")
		return
	}

	desktop, err := h.desktopService.GetDesktop(c.Request.Context(), userID.(int64), desktopID)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此设备")
		default:
			response.InternalError(c, "获取设备信息失败")
		}
		return
	}

	response.Success(c, desktop)
}

// UpdateDesktop 更新设备信息
// @Summary 更新设备信息
// @Description 更新指定设备的名称、类型等信息
// @Tags 设备
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "设备ID"
// @Param body body service.UpdateDesktopRequest true "要更新的字段"
// @Success 200 {object} response.Response{data=service.DesktopResponse}
// @Router /api/desktops/{id} [put]
func (h *DesktopHandler) UpdateDesktop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	desktopID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的设备ID")
		return
	}

	var req service.UpdateDesktopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	desktop, err := h.desktopService.UpdateDesktop(c.Request.Context(), userID.(int64), desktopID, &req)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权修改此设备")
		default:
			response.InternalError(c, "更新设备信息失败")
		}
		return
	}

	response.Success(c, desktop)
}

// DeleteDesktop 删除设备
// @Summary 删除设备
// @Description 删除指定的设备（会同时删除关联的会话和消息）
// @Tags 设备
// @Security Bearer
// @Produce json
// @Param id path int true "设备ID"
// @Success 204 "删除成功"
// @Router /api/desktops/{id} [delete]
func (h *DesktopHandler) DeleteDesktop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	desktopID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的设备ID")
		return
	}

	err = h.desktopService.DeleteDesktop(c.Request.Context(), userID.(int64), desktopID)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权删除此设备")
		default:
			response.InternalError(c, "删除设备失败")
		}
		return
	}

	response.NoContent(c)
}

// GetDesktopStatus 获取设备在线状态
// @Summary 获取设备在线状态
// @Description 获取指定设备的实时在线状态
// @Tags 设备
// @Security Bearer
// @Produce json
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response{data=object}
// @Router /api/desktops/{id}/status [get]
func (h *DesktopHandler) GetDesktopStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	desktopID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的设备ID")
		return
	}

	desktop, err := h.desktopService.GetDesktop(c.Request.Context(), userID.(int64), desktopID)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此设备")
		default:
			response.InternalError(c, "获取设备状态失败")
		}
		return
	}

	response.Success(c, gin.H{
		"desktop_id": desktop.ID,
		"status":     desktop.Status,
	})
}

// RegisterDesktop 注册并绑定一台桌面设备
// @Summary 注册桌面设备（账号直连）
// @Description 登录后直接在电脑端注册设备，返回 desktop_id 与桌面 token
// @Tags 设备
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body service.RegisterDesktopRequest true "设备信息"
// @Success 200 {object} response.Response{data=object}
// @Router /api/desktops/register [post]
func (h *DesktopHandler) RegisterDesktop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.RegisterDesktopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.desktopService.RegisterDesktop(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		response.InternalError(c, "注册设备失败")
		return
	}

	// 生成桌面专用 JWT，用于 WebSocket 连接
	desktopToken, err := h.jwtService.GenerateDesktopToken(userID.(int64), result.Desktop.ID, result.DeviceToken)
	if err != nil {
		response.InternalError(c, "生成桌面 Token 失败")
		return
	}

	response.Success(c, gin.H{
		"desktop_id":    result.Desktop.ID,
		"desktop_token": desktopToken,
		"name":          result.Desktop.Name,
		"agent_type":    result.Desktop.AgentType,
		"os_info":       result.Desktop.OSInfo,
		"working_dir":   result.Desktop.WorkingDir,
	})
}
