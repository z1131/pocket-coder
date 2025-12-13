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
// @Param desktop_id query int true "设备ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=SessionListResponse}
// @Router /api/v1/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 解析设备 ID（从 query 参数获取）
	desktopID, err := strconv.ParseInt(c.Query("desktop_id"), 10, 64)
	if err != nil || desktopID <= 0 {
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
// @Param body body service.CreateSessionRequest true "会话配置"
// @Success 201 {object} response.Response{data=service.SessionResponse}
// @Router /api/v1/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req service.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求参数")
		return
	}

	if req.DesktopID <= 0 {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	session, err := h.sessionService.CreateSession(c.Request.Context(), userID.(int64), req.DesktopID, &req)
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

	// 关键补充：创建 DB 记录后，必须通知 Desktop Agent 创建真实的 PTY 进程
	// 注意：这里需要 Hub 的引用才能发送 WebSocket 消息。
	// 但 SessionHandler 没有直接引用 Hub。
	// 架构上，通常是在 SessionService 中调用 Hub (通过 Interface) 或者 Event Bus。
	// 
	// 不过，在这个项目中，Hub 依赖 SessionService。如果 SessionService 又依赖 Hub，会造成循环依赖。
	// 
	// 临时解决方案：我们在 SessionService.CreateSession 内部（或者调用它的地方）处理通知。
	// 
	// 我们的 Hub 已经在 `handleUserMessage` 中有类似的逻辑：
	// if session == nil { h.notifyDesktopClient(..., TypeSessionCreate, ...) }
	//
	// 所以，这里我们是否需要主动通知？
	// 是的，因为这是通过 HTTP API 创建的会话，Agent 此时并不知道。
	//
	// 问题：SessionHandler 无法直接访问 WebSocket 连接。
	// 
	// 解决方案：
	// 1. 修改 SessionService，让它有一个 callback 或者 event mechanism。
	// 2. 将 Hub 注入到 Handler (但这违反了 Clean Arch, Handler 应该只调 Service)。
	// 3. 在 Service 层引入一个 `NotificationService` 接口，Hub 实现它，断开循环依赖。
	
	// 鉴于时间，我们这里暂时不修改 Handler 去发通知，而是修改 SessionService。
	// 但 SessionService 里也没有 Hub。
	//
	// 让我们看看 Hub.go。Hub 引用了 SessionService。
	// 
	// 也许我们可以让 SessionService 有一个 `OnSessionCreated` hook?
	// 
	// 或者，更简单的：前端调用这个 API 成功后，Agent 实际上还没有 PTY。
	// 只有当用户真的发送消息时，Agent 才会收到消息。
	// 但我们希望用户一点“新建会话”，终端就弹出来。
	//
	// 回到 `handleUserMessage` in Hub.go:
	// 它会在 sessionID 为空时自动创建 session 并发送 `TypeSessionCreate`。
	//
	// 这里的 `CreateSession` API 主要是为了 UI 上显示“新会话”。
	// 
	// 如果我们只是在 DB 里创建了，Agent 不知道。
	// 当用户在前端进入这个新 Session 页面，尝试发送 Input 时，`handleTerminalInput` 会携带 `session_id`。
	// Agent 收到 `terminal:input` 后，如果发现 `session_id` 对应的 PTY 不存在，是否应该自动创建？
	//
	// 我们在 CLI 的 `HandleSessionCreate` 中写了：
	// "如果接收到此消息且本地没有任何关联的会话，将其绑定到默认 PTY"
	// "否则，启动一个新的后台终端"
	//
	// 所以，如果 Agent 收到一个针对未知 Session ID 的 Input，它可能会困惑。
	//
	// 最好的办法是 Server 必须通知 Agent "Start Session X"。
	//
	// 我将修改 `server/internal/service/session_service.go` 来支持注入一个 `SessionNotifier` 接口。
	
	response.Created(c, session)
}

// GetSession 获取会话详情
// @Summary 获取会话详情
// @Description 获取指定会话的详细信息
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
// @Description 删除指定的会话
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
