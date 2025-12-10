// Package websocket 提供 WebSocket 通信功能
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/service"
)

// Hub 是 WebSocket 连接的中心管理器
// 负责：
// 1. 管理所有客户端连接
// 2. 路由消息
// 3. 同步在线状态
type Hub struct {
	// 手机端客户端映射：userID -> []*Client
	// 一个用户可能有多个手机连接（多设备登录）
	mobileClients map[int64][]*Client

	// 电脑端客户端映射：desktopID -> *Client
	// 一个电脑只有一个连接
	desktopClients map[int64]*Client

	// 用户到设备的映射：userID -> []desktopID
	// 用于快速查找用户的所有设备
	userDesktops map[int64][]int64

	// 注册通道
	register chan *Client

	// 注销通道
	unregister chan *Client

	// 互斥锁，保护并发访问
	mu sync.RWMutex

	// 依赖的服务
	desktopService *service.DesktopService
	sessionService *service.SessionService
	cache          *cache.RedisCache
}

// NewHub 创建 Hub 实例
func NewHub(
	desktopService *service.DesktopService,
	sessionService *service.SessionService,
	cache *cache.RedisCache,
) *Hub {
	return &Hub{
		mobileClients:  make(map[int64][]*Client),
		desktopClients: make(map[int64]*Client),
		userDesktops:   make(map[int64][]int64),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		desktopService: desktopService,
		sessionService: sessionService,
		cache:          cache,
	}
}

// Run 启动 Hub 的主循环
// 应该在单独的 goroutine 中运行
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch client.clientType {
	case ClientTypeMobile:
		// 添加到手机端列表
		h.mobileClients[client.userID] = append(h.mobileClients[client.userID], client)
		log.Printf("Mobile client registered: userID=%d", client.userID)

	case ClientTypeDesktop:
		// 检查是否已有连接（替换旧连接）
		if old, exists := h.desktopClients[client.desktopID]; exists {
			old.Close()
		}

		// 添加到电脑端映射
		h.desktopClients[client.desktopID] = client

		// 更新用户到设备的映射
		h.updateUserDesktops(client.userID, client.desktopID, true)

		// 更新 Redis 在线状态
		go func() {
			ctx := context.Background()
			if err := h.desktopService.SetDesktopOnline(ctx, client.desktopID, client.userID); err != nil {
				log.Printf("Failed to set desktop online: %v", err)
			}

			// 通知用户的手机端设备上线
			h.notifyMobileClients(client.userID, NewMessage(TypeDesktopOnline, &DesktopStatusPayload{
				DesktopID: client.desktopID,
			}))
		}()

		log.Printf("Desktop client registered: desktopID=%d, userID=%d", client.desktopID, client.userID)
	}
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch client.clientType {
	case ClientTypeMobile:
		// 从手机端列表移除
		clients := h.mobileClients[client.userID]
		for i, c := range clients {
			if c == client {
				h.mobileClients[client.userID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		// 如果没有连接了，删除 key
		if len(h.mobileClients[client.userID]) == 0 {
			delete(h.mobileClients, client.userID)
		}
		log.Printf("Mobile client unregistered: userID=%d", client.userID)

	case ClientTypeDesktop:
		// 检查是否是当前连接
		if current, exists := h.desktopClients[client.desktopID]; exists && current == client {
			delete(h.desktopClients, client.desktopID)

			// 更新用户到设备的映射
			h.updateUserDesktops(client.userID, client.desktopID, false)

			// 更新 Redis 离线状态
			go func() {
				ctx := context.Background()
				if err := h.desktopService.SetDesktopOffline(ctx, client.desktopID, client.userID); err != nil {
					log.Printf("Failed to set desktop offline: %v", err)
				}

				// 通知用户的手机端设备下线
				h.notifyMobileClients(client.userID, NewMessage(TypeDesktopOffline, &DesktopStatusPayload{
					DesktopID: client.desktopID,
				}))
			}()

			log.Printf("Desktop client unregistered: desktopID=%d, userID=%d", client.desktopID, client.userID)
		}
	}

	// 关闭客户端
	client.Close()
}

// updateUserDesktops 更新用户到设备的映射
func (h *Hub) updateUserDesktops(userID, desktopID int64, add bool) {
	desktops := h.userDesktops[userID]

	if add {
		// 添加（去重）
		for _, id := range desktops {
			if id == desktopID {
				return
			}
		}
		h.userDesktops[userID] = append(desktops, desktopID)
	} else {
		// 移除
		for i, id := range desktops {
			if id == desktopID {
				h.userDesktops[userID] = append(desktops[:i], desktops[i+1:]...)
				break
			}
		}
		if len(h.userDesktops[userID]) == 0 {
			delete(h.userDesktops, userID)
		}
	}
}

// notifyMobileClients 向用户的所有手机端发送消息
func (h *Hub) notifyMobileClients(userID int64, msg *Message) {
	h.mu.RLock()
	clients := h.mobileClients[userID]
	h.mu.RUnlock()

	for _, client := range clients {
		client.SendMessage(msg)
	}
}

// notifyDesktopClient 向电脑端发送消息
func (h *Hub) notifyDesktopClient(desktopID int64, msg *Message) bool {
	h.mu.RLock()
	client, exists := h.desktopClients[desktopID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	client.SendMessage(msg)
	return true
}

// Register 注册客户端（供外部调用）
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端（供外部调用）
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// handleHeartbeat 处理心跳消息
func (h *Hub) handleHeartbeat(client *Client) {
	if client.clientType != ClientTypeDesktop {
		return
	}

	// 更新 Redis 心跳
	go func() {
		ctx := context.Background()
		if err := h.desktopService.UpdateHeartbeat(ctx, client.desktopID); err != nil {
			log.Printf("Failed to update heartbeat: %v", err)
		}
	}()
}

// handleUserMessage 处理用户消息（手机端 → 电脑端）
func (h *Hub) handleUserMessage(client *Client, msg *Message) {
	// 解析 Payload
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid user message payload")
		return
	}

	// 提取参数
	desktopID := int64(payload["desktop_id"].(float64))
	content := payload["content"].(string)

	var sessionID int64
	if sid, ok := payload["session_id"].(float64); ok {
		sessionID = int64(sid)
	}

	ctx := context.Background()

	// 检查设备是否在线
	if !h.desktopService.IsDesktopOnline(ctx, desktopID) {
		client.SendMessage(NewMessage(TypeError, &ErrorPayload{
			Code:    1202,
			Message: "设备已离线",
		}))
		return
	}

	// 检查设备所有权
	desktop, err := h.desktopService.GetDesktopByID(ctx, desktopID)
	if err != nil || desktop == nil || desktop.UserID != client.userID {
		client.SendMessage(NewMessage(TypeError, &ErrorPayload{
			Code:    1003,
			Message: "无权操作此设备",
		}))
		return
	}

	// 如果没有指定会话，获取或创建活跃会话
	if sessionID == 0 {
		session, err := h.sessionService.GetActiveSession(ctx, client.userID, desktopID)
		if err != nil {
			log.Printf("Failed to get active session: %v", err)
		}
		if session != nil {
			sessionID = session.ID
		} else {
			// 创建新会话
			newSession, err := h.sessionService.CreateSession(ctx, client.userID, desktopID, nil)
			if err != nil {
				client.SendMessage(NewMessage(TypeError, &ErrorPayload{
					Code:    500,
					Message: "创建会话失败",
				}))
				return
			}
			sessionID = newSession.ID

			// 通知电脑端创建会话
			h.notifyDesktopClient(desktopID, NewMessage(TypeSessionCreate, &SessionCreatePayload{
				SessionID: sessionID,
			}))
		}
	}

	// 保存用户消息到数据库
	_, err = h.sessionService.AddMessage(ctx, sessionID, model.MessageRoleUser, content)
	if err != nil {
		log.Printf("Failed to save user message: %v", err)
	}

	// 转发消息给电脑端
	h.notifyDesktopClient(desktopID, NewMessageWithID(TypeUserMessage, &UserMessagePayload{
		DesktopID: desktopID,
		SessionID: sessionID,
		Content:   content,
	}, msg.MessageID))
}

// handleAgentResponse 处理 AI 完整响应（电脑端 → 手机端）
func (h *Hub) handleAgentResponse(client *Client, msg *Message) {
	// 解析 Payload
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload AgentResponsePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Printf("Invalid agent response payload: %v", err)
		return
	}

	ctx := context.Background()

	// 保存 AI 响应到数据库
	_, err := h.sessionService.AddMessage(ctx, payload.SessionID, model.MessageRoleAssistant, payload.Content)
	if err != nil {
		log.Printf("Failed to save agent response: %v", err)
	}

	// 转发给用户的手机端
	h.notifyMobileClients(client.userID, msg)
}

// handleAgentStream 处理 AI 流式输出（电脑端 → 手机端）
func (h *Hub) handleAgentStream(client *Client, msg *Message) {
	// 直接转发给用户的手机端
	h.notifyMobileClients(client.userID, msg)
}

// handleAgentStatus 处理 AI 状态变更（电脑端 → 手机端）
func (h *Hub) handleAgentStatus(client *Client, msg *Message) {
	// 直接转发给用户的手机端
	h.notifyMobileClients(client.userID, msg)
}

// handleTerminalToDesktop 处理终端消息（手机端 → 电脑端）
func (h *Hub) handleTerminalToDesktop(client *Client, msg *Message) {
	// 从 payload 获取目标设备 ID
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid terminal message payload")
		return
	}

	desktopID, ok := payload["desktop_id"].(float64)
	if !ok || desktopID == 0 {
		// 如果没有指定 desktop_id，尝试发给用户的第一个在线设备
		onlineDesktops := h.GetOnlineDesktops(client.userID)
		if len(onlineDesktops) == 0 {
			client.SendMessage(NewMessage(TypeError, &ErrorPayload{
				Code:    1202,
				Message: "没有在线的设备",
			}))
			return
		}
		desktopID = float64(onlineDesktops[0])
	}

	// 检查设备所有权
	ctx := context.Background()
	desktop, err := h.desktopService.GetDesktopByID(ctx, int64(desktopID))
	if err != nil || desktop == nil || desktop.UserID != client.userID {
		client.SendMessage(NewMessage(TypeError, &ErrorPayload{
			Code:    1003,
			Message: "无权操作此设备",
		}))
		return
	}

	// 转发给电脑端
	log.Printf("Forwarding terminal message to desktop %d: %v", desktopID, payload)
	if h.notifyDesktopClient(int64(desktopID), msg) {
		log.Printf("Successfully forwarded to desktop %d", desktopID)
	} else {
		log.Printf("Failed to forward to desktop %d: client not connected", desktopID)
	}
}

// handleTerminalToMobile 处理终端消息（电脑端 → 手机端）
func (h *Hub) handleTerminalToMobile(client *Client, msg *Message) {
	// 直接转发给用户的所有手机端
	h.notifyMobileClients(client.userID, msg)
}

// IsDesktopConnected 检查设备是否已连接
func (h *Hub) IsDesktopConnected(desktopID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.desktopClients[desktopID]
	return exists
}

// GetOnlineDesktops 获取用户的在线设备
func (h *Hub) GetOnlineDesktops(userID int64) []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.userDesktops[userID]
}
