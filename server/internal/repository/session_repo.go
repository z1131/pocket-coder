// Package repository 提供数据访问层的实现
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pocket-coder-server/internal/model"
)

// SessionRepository 会话数据访问层
// 负责会话相关的所有数据库操作
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository 创建 SessionRepository 实例
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 创建新会话
// 参数:
//   - ctx: 上下文
//   - session: 会话对象，ID 和时间字段会被自动填充
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID 根据 ID 获取会话
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//
// 返回:
//   - *model.Session: 会话对象，未找到返回 nil
//   - error: 数据库错误
func (r *SessionRepository) GetByID(ctx context.Context, id int64) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetByIDWithMessages 根据 ID 获取会话及其所有消息
// 用于加载会话历史
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//
// 返回:
//   - *model.Session: 包含 Messages 字段的会话对象
//   - error: 数据库错误
func (r *SessionRepository) GetByIDWithMessages(ctx context.Context, id int64) (*model.Session, error) {
	var session model.Session
	// Preload 预加载消息，并按创建时间排序
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC") // 按时间正序，最早的在前
		}).
		First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetByIDWithDesktop 根据 ID 获取会话及其所属设备
// 用于验证会话所有权
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//
// 返回:
//   - *model.Session: 包含 Desktop 字段的会话对象
//   - error: 数据库错误
func (r *SessionRepository) GetByIDWithDesktop(ctx context.Context, id int64) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).Preload("Desktop").First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetByDesktopID 获取设备的所有会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - []model.Session: 会话列表，按创建时间倒序
//   - error: 数据库错误
func (r *SessionRepository) GetByDesktopID(ctx context.Context, desktopID int64) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.WithContext(ctx).
		Where("desktop_id = ?", desktopID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetByDesktopIDWithPagination 分页获取设备的会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - page: 页码，从 1 开始
//   - pageSize: 每页数量
//
// 返回:
//   - []model.Session: 会话列表
//   - int64: 总数量（用于计算总页数）
//   - error: 数据库错误
func (r *SessionRepository) GetByDesktopIDWithPagination(ctx context.Context, desktopID int64, page, pageSize int) ([]model.Session, int64, error) {
	var sessions []model.Session
	var total int64

	// 构建基础查询
	query := r.db.WithContext(ctx).Model(&model.Session{}).Where("desktop_id = ?", desktopID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	// Offset: 跳过的记录数 = (页码 - 1) * 每页数量
	// Limit: 每页返回的最大记录数
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}

// GetActiveByDesktopID 获取设备当前活跃的会话
// 通常一个设备只有一个活跃会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - *model.Session: 活跃会话，如果没有返回 nil
//   - error: 数据库错误
func (r *SessionRepository) GetActiveByDesktopID(ctx context.Context, desktopID int64) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).
		Where("desktop_id = ? AND status = ?", desktopID, model.SessionStatusActive).
		Order("created_at DESC"). // 取最新的
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// Update 更新会话信息
// 参数:
//   - ctx: 上下文
//   - session: 包含要更新字段的会话对象，必须包含 ID
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// EndSession 结束会话
// 将会话状态设为 ended，并记录结束时间
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) EndSession(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   model.SessionStatusEnded,
			"ended_at": now,
		}).Error
}

// EndAllActiveByDesktopID 结束设备的所有活跃会话
// 当设备离线或用户开始新会话时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) EndAllActiveByDesktopID(ctx context.Context, desktopID int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("desktop_id = ? AND status = ?", desktopID, model.SessionStatusActive).
		Updates(map[string]interface{}{
			"status":   model.SessionStatusEnded,
			"ended_at": now,
		}).Error
}

// Delete 删除会话
// 注意: 会级联删除关联的所有消息
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Session{}, id).Error
}

// CountByDesktopID 统计设备的会话数量
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - int64: 会话数量
//   - error: 数据库错误
func (r *SessionRepository) CountByDesktopID(ctx context.Context, desktopID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Session{}).Where("desktop_id = ?", desktopID).Count(&count).Error
	return count, err
}

// UpdateSummary 更新会话的标题和摘要
// 参数:
//   - ctx: 上下文
//   - id: 会话ID
//   - title: 标题
//   - summary: 摘要
//
// 返回:
//   - error: 数据库错误
func (r *SessionRepository) UpdateSummary(ctx context.Context, id int64, title, summary string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"title":   title,
			"summary": summary,
		}).Error
}
