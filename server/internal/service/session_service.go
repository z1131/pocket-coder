// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
)

// SessionNotifier 会话通知接口
type SessionNotifier interface {
	NotifySessionCreate(desktopID int64, sessionID int64, workingDir string, isDefault bool)
	NotifySessionClose(desktopID int64, sessionID int64)
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
	desktopRepo *repository.DesktopRepository // 设备数据访问层
	cache       *cache.RedisCache             // Redis 缓存
	notifier    SessionNotifier               // 会话通知器
}

// NewSessionService 创建 SessionService 实例
func NewSessionService(
	sessionRepo *repository.SessionRepository,
	desktopRepo *repository.DesktopRepository,
	cache *cache.RedisCache,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		desktopRepo: desktopRepo,
		cache:       cache,
	}
}

// SetNotifier 设置通知器
func (s *SessionService) SetNotifier(n SessionNotifier) {
	s.notifier = n
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID        int64   `json:"id"`
	DesktopID int64   `json:"desktop_id"`
	AgentType string  `json:"agent_type"`
	IsDefault bool    `json:"is_default"` // 是否为默认会话
	WorkingDir *string `json:"working_dir,omitempty"`
	Title     *string `json:"title,omitempty"`
	Preview   *string `json:"preview,omitempty"` // Base64 编码的最近输出
	Status    string  `json:"status"`
	StartedAt string  `json:"started_at"`
	EndedAt   *string `json:"ended_at,omitempty"`
}

// SessionDetailResponse 会话详情响应
type SessionDetailResponse struct {
	Session SessionResponse `json:"session"`
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	DesktopID  int64   `json:"desktop_id"`             // 设备ID
	WorkingDir *string `json:"working_dir"`            // 工作目录（可选）
	IsDefault *bool   `json:"is_default" json:"-"`          // 是否为默认会话（由服务端控制）
}

// CreateSession 创建新会话
func (s *SessionService) CreateSession(ctx context.Context, userID, desktopID int64, req *CreateSessionRequest) (*SessionResponse, error) {
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

	// 手机端 API 创建的会话默认都是非默认会话
	isDefault := false
	// 如果是 EnsureDefaultSession 调用，会通过 req.IsDefault 传入 true
	if req != nil && req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	session := &model.Session{
		DesktopID: desktopID,
		AgentType: "claude-code", // 默认值，后续可由 Client 指定
		Status:    model.SessionStatusActive,
		IsDefault: isDefault,
	}
	if req != nil && req.WorkingDir != nil {
		session.WorkingDir = req.WorkingDir
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	if err := s.cache.SetActiveSession(ctx, desktopID, session.ID); err != nil {
		// Non-fatal error
	}

	if s.notifier != nil {
		wd := ""
		if session.WorkingDir != nil {
			wd = *session.WorkingDir
		}
		go s.notifier.NotifySessionCreate(desktopID, session.ID, wd, isDefault)
	}

	return s.toSessionResponse(session), nil
}

// ListSessions 获取设备的会话列表
func (s *SessionService) ListSessions(ctx context.Context, userID, desktopID int64, page, pageSize int) ([]SessionResponse, int64, error) {
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

	sessions, total, err := s.sessionRepo.GetByDesktopIDWithPagination(ctx, desktopID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]SessionResponse, len(sessions))
	for i, session := range sessions {
		resp := s.toSessionResponse(&session)

		// 获取预览数据（最近 1KB）
		tail, err := s.cache.GetTerminalHistoryTail(ctx, session.ID, 1024)
		if err == nil && len(tail) > 0 {
			encoded := base64.StdEncoding.EncodeToString(tail)
			resp.Preview = &encoded
		}
		
		result[i] = *resp
	}
	return result, total, nil
}

// GetSession 获取会话详情
func (s *SessionService) GetSession(ctx context.Context, userID, sessionID int64) (*SessionDetailResponse, error) {
	session, err := s.sessionRepo.GetByIDWithDesktop(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.Desktop == nil || session.Desktop.UserID != userID {
		return nil, ErrNoPermission
	}

	return &SessionDetailResponse{
		Session: *s.toSessionResponse(session),
	}, nil
}

// GetActiveSession 获取设备的当前活跃会话
func (s *SessionService) GetActiveSession(ctx context.Context, userID, desktopID int64) (*SessionResponse, error) {
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

	sessionID, err := s.cache.GetActiveSession(ctx, desktopID)
	if err != nil {
		return nil, err
	}

	var session *model.Session
	if sessionID > 0 {
		session, err = s.sessionRepo.GetByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}
	}
	if session == nil {
		session, err = s.sessionRepo.GetActiveByDesktopID(ctx, desktopID)
		if err != nil {
			return nil, err
		}
	}
	if session == nil {
		return nil, nil
	}
	return s.toSessionResponse(session), nil
}

// ListActiveSessions 获取设备的所有活跃会话（用于 CLI 重连恢复）
func (s *SessionService) ListActiveSessions(ctx context.Context, desktopID int64) ([]*model.Session, error) {
	sessions, err := s.sessionRepo.GetAllActiveByDesktopID(ctx, desktopID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Session, len(sessions))
	for i := range sessions {
		result[i] = &sessions[i]
	}
	return result, nil
}

// EndSession 结束会话 (软删除 + 通知关闭 + 归档日志)
func (s *SessionService) EndSession(ctx context.Context, userID, sessionID int64) error {
	session, err := s.sessionRepo.GetByIDWithDesktop(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}
	if session.Desktop == nil || session.Desktop.UserID != userID {
		return ErrNoPermission
	}
	if session.Status == model.SessionStatusEnded {
		return nil // 已经结束，无需重复操作
	}

	// 1. 更新数据库状态为 ended
	if err := s.sessionRepo.EndSession(ctx, sessionID); err != nil {
		return err
	}

	// 2. 清除 Redis 活跃会话状态
	_ = s.cache.ClearActiveSession(ctx, session.DesktopID)

	// 3. 异步归档日志并通知 CLI 关闭
	go func() {
		// 通知 CLI 关闭终端
		if s.notifier != nil {
			s.notifier.NotifySessionClose(session.DesktopID, sessionID)
		}

		// 归档日志：从 Redis 读取并存储到 LogDump
		history, err := s.cache.GetTerminalHistory(ctx, sessionID)
		if err == nil && len(history) > 0 {
			logContent := string(history)
			_ = s.sessionRepo.UpdateLogDump(ctx, sessionID, logContent) // 存储到数据库
		}
		_ = s.cache.ClearTerminalHistory(ctx, sessionID) // 清除 Redis 历史
	}()

	return nil
}

// DeleteSession 删除会话 (改为调用 EndSession，实现软删除)
func (s *SessionService) DeleteSession(ctx context.Context, userID, sessionID int64) error {
	// 实际上是结束会话，而不是硬删除
	return s.EndSession(ctx, userID, sessionID)
}

// GetSessionByID 获取会话（内部使用，不验证权限）
func (s *SessionService) GetSessionByID(ctx context.Context, sessionID int64) (*model.Session, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// toSessionResponse 将会话模型转换为响应格式
func (s *SessionService) toSessionResponse(session *model.Session) *SessionResponse {
	resp := &SessionResponse{
		ID:        session.ID,
		DesktopID: session.DesktopID,
		AgentType: session.AgentType,
		IsDefault: session.IsDefault, // 使用 IsDefault
		WorkingDir: session.WorkingDir,
		Title:     session.Title,
		Status:    session.Status,
		StartedAt: session.StartedAt.Format(time.RFC3339),
	}
	if session.EndedAt != nil {
		formatted := session.EndedAt.Format(time.RFC3339)
		resp.EndedAt = &formatted
	}
	return resp
}

// EnsureDefaultSession 确保设备有活跃的默认会话（用于 Agent 连入时）
func (s *SessionService) EnsureDefaultSession(ctx context.Context, userID, desktopID int64) (*SessionResponse, error) {
	// 1. 查找是否存在活跃的默认会话
	activeDefaultSession, err := s.sessionRepo.GetActiveDefaultSessionByDesktopID(ctx, desktopID) // 需要实现这个方法
	if err != nil {
		return nil, err
	}
	if activeDefaultSession != nil {
		return s.toSessionResponse(activeDefaultSession), nil
	}

	// 2. 如果不存在，则创建一个新的默认会话
	isDefault := true
	return s.CreateSession(ctx, userID, desktopID, &CreateSessionRequest{
		DesktopID: desktopID,
		IsDefault: &isDefault,
	})
}

// ResetSessions 重置设备的所有会话（用于 CLI 重启时清理旧会话）
func (s *SessionService) ResetSessions(ctx context.Context, desktopID int64) error {
	// 1. 结束所有活跃会话
	if err := s.sessionRepo.EndAllActiveByDesktopID(ctx, desktopID); err != nil {
		return err
	}
	// 2. 清除 Redis 活跃会话状态
	return s.cache.ClearActiveSession(ctx, desktopID)
}