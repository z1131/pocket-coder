// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"errors"
	"time"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
)

// SessionNotifier 会话通知接口
type SessionNotifier interface {
	NotifySessionCreate(desktopID int64, sessionID int64, workingDir string)
}

// 会话服务相关错误
var (
	ErrSessionNotFound = errors.New("会话不存在")
	ErrSessionEnded    = errors.New("会话已结束")
)

// SessionService 会话服务
// 处理用户与 AI 的对话会话
type SessionService struct {
	sessionRepo *repository.SessionRepository // 会话数据访问层
	messageRepo *repository.MessageRepository // 消息数据访问层
	desktopRepo *repository.DesktopRepository // 设备数据访问层
	cache       *cache.RedisCache             // Redis 缓存
	aiService   *AIService                    // AI 服务（可选）
	notifier    SessionNotifier               // 会话通知器
}

// NewSessionService 创建 SessionService 实例
func NewSessionService(
	sessionRepo *repository.SessionRepository,
	messageRepo *repository.MessageRepository,
	desktopRepo *repository.DesktopRepository,
	cache *cache.RedisCache,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		desktopRepo: desktopRepo,
		cache:       cache,
	}
}

// SetNotifier 设置通知器
func (s *SessionService) SetNotifier(n SessionNotifier) {
	s.notifier = n
}

// SetAIService 设置 AI 服务
func (s *SessionService) SetAIService(aiService *AIService) {
	s.aiService = aiService
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID         int64   `json:"id"`
	DesktopID  int64   `json:"desktop_id"`
	AgentType  string  `json:"agent_type"`
	WorkingDir *string `json:"working_dir,omitempty"`
	Title      *string `json:"title,omitempty"`
	Summary    *string `json:"summary,omitempty"`
	Status     string  `json:"status"`
	StartedAt  string  `json:"started_at"`
	EndedAt    *string `json:"ended_at,omitempty"`
}

// SessionDetailResponse 会话详情响应（包含消息）
type SessionDetailResponse struct {
	Session  SessionResponse   `json:"session"`
	Messages []MessageResponse `json:"messages"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID        int64  `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	DesktopID  int64   `json:"desktop_id"`  // 设备ID
	WorkingDir *string `json:"working_dir"` // 工作目录（可选）
}

// CreateSession 创建新会话
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//   - req: 创建请求
//
// 返回:
//   - *SessionResponse: 创建的会话
//   - error: 操作错误
func (s *SessionService) CreateSession(ctx context.Context, userID, desktopID int64, req *CreateSessionRequest) (*SessionResponse, error) {
	// 1. 获取设备信息
	desktop, err := s.desktopRepo.GetByID(ctx, desktopID)
	if err != nil {
		return nil, err
	}
	if desktop == nil {
		return nil, ErrDesktopNotFound
	}

	// 2. 验证权限
	if desktop.UserID != userID {
		return nil, ErrNoPermission
	}

	// 3. (已移除) 不再强制结束之前的活跃会话，支持多会话并存
	// if err := s.sessionRepo.EndAllActiveByDesktopID(ctx, desktopID); err != nil {
	// 	return nil, err
	// }

	// 4. 创建新会话
	session := &model.Session{
		DesktopID: desktopID,
		AgentType: desktop.AgentType,
		Status:    model.SessionStatusActive,
	}

	// 设置工作目录：优先使用请求中的，否则使用设备的
	if req != nil && req.WorkingDir != nil {
		session.WorkingDir = req.WorkingDir
	} else {
		session.WorkingDir = desktop.WorkingDir
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// 5. 更新 Redis 中的活跃会话
	if err := s.cache.SetActiveSession(ctx, desktopID, session.ID); err != nil {
		// 非致命错误，记录日志但不返回错误
	}

	// 6. 通知 Agent 创建会话
	if s.notifier != nil {
		wd := ""
		if session.WorkingDir != nil {
			wd = *session.WorkingDir
		}
		// 异步通知，避免阻塞
		go s.notifier.NotifySessionCreate(desktopID, session.ID, wd)
	}

	return s.toSessionResponse(session), nil
}

// ListSessions 获取设备的会话列表
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//   - page: 页码（从 1 开始）
//   - pageSize: 每页数量
//
// 返回:
//   - []SessionResponse: 会话列表
//   - int64: 总数量
//   - error: 操作错误
func (s *SessionService) ListSessions(ctx context.Context, userID, desktopID int64, page, pageSize int) ([]SessionResponse, int64, error) {
	// 1. 获取设备并验证权限
	desktop, err := s.desktopRepo.GetByID(ctx, desktopID)
	if err != nil {
		return nil, 0, err
	}
	if desktop == nil {
		return nil, 0, ErrDesktopNotFound
	}
	if desktop.UserID != userID {
		return nil, 0, ErrNoPermission
	}

	// 2. 分页获取会话
	sessions, total, err := s.sessionRepo.GetByDesktopIDWithPagination(ctx, desktopID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 3. 转换为响应格式
	result := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		result[i] = *s.toSessionResponse(&session)
	}

	return result, total, nil
}

// GetSession 获取会话详情（包含消息）
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - sessionID: 会话ID
//
// 返回:
//   - *SessionDetailResponse: 会话详情（包含消息历史）
//   - error: 操作错误
func (s *SessionService) GetSession(ctx context.Context, userID, sessionID int64) (*SessionDetailResponse, error) {
	// 1. 获取会话（包含设备信息）
	session, err := s.sessionRepo.GetByIDWithDesktop(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// 2. 验证权限
	if session.Desktop == nil || session.Desktop.UserID != userID {
		return nil, ErrNoPermission
	}

	// 3. 获取消息历史
	messages, err := s.messageRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 4. 构建响应
	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = MessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		}
	}

	return &SessionDetailResponse{
		Session:  *s.toSessionResponse(session),
		Messages: messageResponses,
	}, nil
}

// GetActiveSession 获取设备的当前活跃会话
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//
// 返回:
//   - *SessionResponse: 活跃会话，如果没有返回 nil
//   - error: 操作错误
func (s *SessionService) GetActiveSession(ctx context.Context, userID, desktopID int64) (*SessionResponse, error) {
	// 1. 验证权限
	desktop, err := s.desktopRepo.GetByID(ctx, desktopID)
	if err != nil {
		return nil, err
	}
	if desktop == nil {
		return nil, ErrDesktopNotFound
	}
	if desktop.UserID != userID {
		return nil, ErrNoPermission
	}

	// 2. 先从 Redis 缓存获取
	sessionID, err := s.cache.GetActiveSession(ctx, desktopID)
	if err != nil {
		return nil, err
	}

	var session *model.Session
	if sessionID > 0 {
		// 从数据库获取会话详情
		session, err = s.sessionRepo.GetByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}
	}

	// 3. 如果缓存没有，从数据库查询
	if session == nil {
		session, err = s.sessionRepo.GetActiveByDesktopID(ctx, desktopID)
		if err != nil {
			return nil, err
		}
	}

	if session == nil {
		return nil, nil // 没有活跃会话
	}

	return s.toSessionResponse(session), nil
}

// EndSession 结束会话
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - sessionID: 会话ID
//
// 返回:
//   - error: 操作错误
func (s *SessionService) EndSession(ctx context.Context, userID, sessionID int64) error {
	// 1. 获取会话
	session, err := s.sessionRepo.GetByIDWithDesktop(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	// 2. 验证权限
	if session.Desktop == nil || session.Desktop.UserID != userID {
		return ErrNoPermission
	}

	// 3. 如果已经结束，直接返回
	if session.Status == model.SessionStatusEnded {
		return nil
	}

	// 4. 结束会话
	if err := s.sessionRepo.EndSession(ctx, sessionID); err != nil {
		return err
	}

	// 5. 异步生成 AI 总结（不阻塞返回）
	go func() {
		if err := s.GenerateSummary(context.Background(), sessionID); err != nil {
			// 记录错误但不影响主流程
			// log.Printf("Failed to generate session summary: %v", err)
		}
	}()

	// 6. 清除 Redis 缓存中的活跃会话
	return s.cache.ClearActiveSession(ctx, session.DesktopID)
}

// DeleteSession 删除会话
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - sessionID: 会话ID
//
// 返回:
//   - error: 操作错误
func (s *SessionService) DeleteSession(ctx context.Context, userID, sessionID int64) error {
	// 1. 获取会话
	session, err := s.sessionRepo.GetByIDWithDesktop(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	// 2. 验证权限
	if session.Desktop == nil || session.Desktop.UserID != userID {
		return ErrNoPermission
	}

	// 3. 如果是活跃会话，先清除缓存
	if session.Status == model.SessionStatusActive {
		_ = s.cache.ClearActiveSession(ctx, session.DesktopID)
	}

	// 4. 删除会话（级联删除消息）
	return s.sessionRepo.Delete(ctx, sessionID)
}

// AddMessage 添加消息到会话
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//   - role: 消息角色（user/assistant/system）
//   - content: 消息内容
//
// 返回:
//   - *model.Message: 创建的消息
//   - error: 操作错误
func (s *SessionService) AddMessage(ctx context.Context, sessionID int64, role, content string) (*model.Message, error) {
	// 1. 检查会话是否存在
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// 2. 检查会话是否已结束
	if session.Status == model.SessionStatusEnded {
		return nil, ErrSessionEnded
	}

	// 3. 创建消息
	message := &model.Message{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetSessionByID 获取会话（内部使用，不验证权限）
func (s *SessionService) GetSessionByID(ctx context.Context, sessionID int64) (*model.Session, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// GetMessages 获取会话消息列表
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - sessionID: 会话ID
//
// 返回:
//   - []MessageResponse: 消息列表
//   - error: 获取失败返回错误
func (s *SessionService) GetMessages(ctx context.Context, userID, sessionID int64) ([]MessageResponse, error) {
	// 1. 获取会话
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// 2. 通过 Desktop 验证权限
	desktop, err := s.desktopRepo.GetByID(ctx, session.DesktopID)
	if err != nil {
		return nil, err
	}
	if desktop == nil || desktop.UserID != userID {
		return nil, ErrNoPermission
	}

	// 3. 获取消息列表
	messages, err := s.messageRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 4. 转换为响应格式
	result := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		result[i] = MessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		}
	}

	return result, nil
}

// toSessionResponse 将会话模型转换为响应格式
func (s *SessionService) toSessionResponse(session *model.Session) *SessionResponse {
	resp := &SessionResponse{
		ID:         session.ID,
		DesktopID:  session.DesktopID,
		AgentType:  session.AgentType,
		WorkingDir: session.WorkingDir,
		Title:      session.Title,
		Summary:    session.Summary,
		Status:     session.Status,
		StartedAt:  session.StartedAt.Format(time.RFC3339),
	}

	if session.EndedAt != nil {
		formatted := session.EndedAt.Format(time.RFC3339)
		resp.EndedAt = &formatted
	}

	return resp
}

// GenerateSummary 为会话生成 AI 总结
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - error: 操作错误
func (s *SessionService) GenerateSummary(ctx context.Context, sessionID int64) error {
	if s.aiService == nil {
		return nil // AI 服务未配置，跳过
	}

	// 获取会话消息
	messages, err := s.messageRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil // 没有消息，跳过
	}

	// 转换为 AI 服务需要的格式
	msgList := make([]map[string]string, len(messages))
	for i, msg := range messages {
		msgList[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// 调用 AI 生成总结
	summary, err := s.aiService.GenerateSessionSummary(ctx, msgList)
	if err != nil {
		return err
	}

	// 更新会话
	return s.sessionRepo.UpdateSummary(ctx, sessionID, summary.Title, summary.Summary)
}

// EnsureDefaultSession 确保设备有活跃会话（用于 Agent 连入时）
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//   - desktopID: 设备ID
//
// 返回:
//   - *SessionResponse: 活跃会话
//   - error: 操作错误
func (s *SessionService) EnsureDefaultSession(ctx context.Context, userID, desktopID int64) (*SessionResponse, error) {
	// 1. 尝试获取现有的活跃会话
	session, err := s.GetActiveSession(ctx, userID, desktopID)
	if err != nil {
		return nil, err
	}

	if session != nil {
		return session, nil
	}

	// 2. 如果没有，创建一个新的
	return s.CreateSession(ctx, userID, desktopID, nil)
}
