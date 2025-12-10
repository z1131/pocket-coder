// Package handler 提供 HTTP 请求处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/pkg/response"
)

// SessionHandler 会话请求处理器
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler 创建 SessionHandler 实例
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// ListSessions 获取设备的会话列表
// @Summary 获取会话列表
// @Description 获取指定设备的所有会话
// @Tags 会话
// @Security Bearer
// @Produce json
// @Param id path int true "设备ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=SessionListResponse}
// @Router /api/desktops/{id}/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 解析设备 ID
	desktopID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的设备ID")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sessions, total, err := h.sessionService.ListSessions(c.Request.Context(), userID.(int64), desktopID, page, pageSize)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此设备")
		default:
			response.InternalError(c, "获取会话列表失败")
		}
		return
	}

	response.Success(c, gin.H{
		"sessions":  sessions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []service.SessionResponse `json:"sessions"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// CreateSession 创建新会话
// @Summary 创建会话
// @Description 为指定设备创建新的会话
// @Tags 会话
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "设备ID"
// @Param body body service.CreateSessionRequest false "会话配置"
// @Success 201 {object} response.Response{data=service.SessionResponse}
// @Router /api/desktops/{id}/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
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

	// 请求体是可选的
	var req service.CreateSessionRequest
	_ = c.ShouldBindJSON(&req)

	session, err := h.sessionService.CreateSession(c.Request.Context(), userID.(int64), desktopID, &req)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权操作此设备")
		default:
			response.InternalError(c, "创建会话失败")
		}
		return
	}

	response.Created(c, session)
}

// GetSession 获取会话详情
// @Summary 获取会话详情
// @Description 获取指定会话的详细信息（包含消息历史）
// @Tags 会话
// @Security Bearer
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response{data=service.SessionDetailResponse}
// @Router /api/sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的会话ID")
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), userID.(int64), sessionID)
	if err != nil {
		switch err {
		case service.ErrSessionNotFound:
			response.SessionNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此会话")
		default:
			response.InternalError(c, "获取会话详情失败")
		}
		return
	}

	response.Success(c, session)
}

// DeleteSession 删除会话
// @Summary 删除会话
// @Description 删除指定的会话（会同时删除所有消息）
// @Tags 会话
// @Security Bearer
// @Produce json
// @Param id path int true "会话ID"
// @Success 204 "删除成功"
// @Router /api/sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的会话ID")
		return
	}

	err = h.sessionService.DeleteSession(c.Request.Context(), userID.(int64), sessionID)
	if err != nil {
		switch err {
		case service.ErrSessionNotFound:
			response.SessionNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权删除此会话")
		default:
			response.InternalError(c, "删除会话失败")
		}
		return
	}

	response.NoContent(c)
}

// GetActiveSession 获取设备的活跃会话
// @Summary 获取活跃会话
// @Description 获取指定设备当前的活跃会话
// @Tags 会话
// @Security Bearer
// @Produce json
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response{data=service.SessionResponse}
// @Router /api/desktops/{id}/sessions/active [get]
func (h *SessionHandler) GetActiveSession(c *gin.Context) {
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

	session, err := h.sessionService.GetActiveSession(c.Request.Context(), userID.(int64), desktopID)
	if err != nil {
		switch err {
		case service.ErrDesktopNotFound:
			response.DeviceNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此设备")
		default:
			response.InternalError(c, "获取活跃会话失败")
		}
		return
	}

	if session == nil {
		response.Success(c, gin.H{
			"session": nil,
		})
		return
	}

	response.Success(c, gin.H{
		"session": session,
	})
}

// GetMessages 获取会话消息列表
// @Summary 获取会话消息列表
// @Description 获取指定会话的所有消息
// @Tags 会话
// @Security Bearer
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} response.Response{data=[]service.MessageResponse}
// @Router /api/sessions/{id}/messages [get]
func (h *SessionHandler) GetMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的会话ID")
		return
	}

	messages, err := h.sessionService.GetMessages(c.Request.Context(), userID.(int64), sessionID)
	if err != nil {
		switch err {
		case service.ErrSessionNotFound:
			response.SessionNotFound(c)
		case service.ErrNoPermission:
			response.Forbidden(c, "无权访问此会话")
		default:
			response.InternalError(c, "获取消息列表失败")
		}
		return
	}

	response.Success(c, gin.H{
		"messages": messages,
	})
}
