// Package websocket 提供 WebSocket 通信功能
package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"pocket-coder-server/internal/service"
	pkgJwt "pocket-coder-server/pkg/jwt"
	"pocket-coder-server/pkg/response"
)

// WebSocket 升级器配置
var upgrader = websocket.Upgrader{
	// 读缓冲区大小
	ReadBufferSize: 1024,
	// 写缓冲区大小
	WriteBufferSize: 1024,
	// 检查来源（生产环境应该验证）
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 生产环境需要检查 Origin
		return true
	},
}

// Handler 处理 WebSocket 连接
type Handler struct {
	hub             *Hub
	desktopService  *service.DesktopService
	jwtSecret       string
}

// NewHandler 创建 WebSocket Handler

func NewHandler(hub *Hub, desktopService *service.DesktopService, jwtSecret string) *Handler {
	return &Handler{
		hub:            hub,
		desktopService: desktopService,
		jwtSecret:      jwtSecret,
	}
}

// HandleMobileWS 处理手机端 WebSocket 连接
// 路由: GET /ws/mobile
// 参数: token (query parameter) - JWT token
func (h *Handler) HandleMobileWS(c *gin.Context) {
	// 从 query 参数获取 token
	token := c.Query("token")
	if token == "" {
		response.Unauthorized(c, "需要认证 token")
		return
	}

	// 验证 JWT token
	claims, err := pkgJwt.ParseUserToken(token, h.jwtSecret)
	if err != nil {
		response.Unauthorized(c, "无效的 token")
		return
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建客户端
	client := NewClient(h.hub, conn, ClientTypeMobile, claims.UserID, 0)

	// 注册客户端
	h.hub.Register(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()

	log.Printf("Mobile WebSocket connected: userID=%d", claims.UserID)
}

// HandleDesktopWS 处理电脑端 WebSocket 连接
// 路由: GET /ws/desktop
// 参数: token (query parameter) - JWT token (设备 token)
func (h *Handler) HandleDesktopWS(c *gin.Context) {
	// 从 query 参数获取 token
	token := c.Query("token")
	if token == "" {
		response.Unauthorized(c, "需要认证 token")
		return
	}

	// 验证设备 JWT token
	claims, err := pkgJwt.ParseDeviceToken(token, h.jwtSecret)
	if err != nil {
		response.Unauthorized(c, "无效的设备 token")
		return
	}

	// 校验桌面归属和令牌有效性
	if ok, err := h.desktopService.ValidateDesktopOwnership(c.Request.Context(), claims.DesktopID, claims.DeviceToken, claims.UserID); err != nil || !ok {
		response.Unauthorized(c, "无效的桌面凭证")
		return
	}

	// 检查设备是否已连接
	if h.hub.IsDesktopConnected(claims.DesktopID) {
		response.Fail(c, http.StatusConflict, "设备已在其他位置连接")
		return
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建客户端
	client := NewClient(h.hub, conn, ClientTypeDesktop, claims.UserID, claims.DesktopID)

	// 注册客户端
	h.hub.Register(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()

	log.Printf("Desktop WebSocket connected: desktopID=%d, userID=%d", claims.DesktopID, claims.UserID)
}

// RegisterRoutes 注册 WebSocket 路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// WebSocket 路由不需要中间件（token 在 query 中验证）
	ws := r.Group("/ws")
	{
		// 手机端 WebSocket
		ws.GET("/mobile", h.HandleMobileWS)
		// 电脑端 WebSocket
		ws.GET("/desktop", h.HandleDesktopWS)
	}
}
