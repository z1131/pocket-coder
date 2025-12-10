// Package repository 提供数据访问层的实现
package repository

import (
	"context"

	"gorm.io/gorm"
	"pocket-coder-server/internal/model"
)

// MessageRepository 消息数据访问层
// 负责消息相关的所有数据库操作
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建 MessageRepository 实例
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create 创建新消息
// 参数:
//   - ctx: 上下文
//   - message: 消息对象，ID 和 CreatedAt 会被自动填充
//
// 返回:
//   - error: 数据库错误
func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// CreateBatch 批量创建消息
// 用于一次性导入多条消息
// 参数:
//   - ctx: 上下文
//   - messages: 消息对象切片
//
// 返回:
//   - error: 数据库错误
func (r *MessageRepository) CreateBatch(ctx context.Context, messages []model.Message) error {
	if len(messages) == 0 {
		return nil
	}
	// CreateInBatches 分批插入，避免单次插入过多数据
	// 100 是每批的数量
	return r.db.WithContext(ctx).CreateInBatches(messages, 100).Error
}

// GetBySessionID 获取会话的所有消息
// 按创建时间正序排列（最早的在前）
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - []model.Message: 消息列表
//   - error: 数据库错误
func (r *MessageRepository) GetBySessionID(ctx context.Context, sessionID int64) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC"). // 按时间正序，方便展示对话
		Find(&messages).Error
	return messages, err
}

// GetBySessionIDWithPagination 分页获取会话的消息
// 用于加载更多历史消息
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//   - page: 页码，从 1 开始
//   - pageSize: 每页数量
//
// 返回:
//   - []model.Message: 消息列表
//   - int64: 总数量
//   - error: 数据库错误
func (r *MessageRepository) GetBySessionIDWithPagination(ctx context.Context, sessionID int64, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Message{}).Where("session_id = ?", sessionID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Order("created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	return messages, total, err
}

// GetLatestBySessionID 获取会话的最新 N 条消息
// 用于显示最近的对话上下文
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//   - limit: 要获取的消息数量
//
// 返回:
//   - []model.Message: 消息列表（按时间正序）
//   - error: 数据库错误
func (r *MessageRepository) GetLatestBySessionID(ctx context.Context, sessionID int64, limit int) ([]model.Message, error) {
	var messages []model.Message

	// 子查询：先按时间倒序取最新的 N 条
	// 然后外层查询再按时间正序排列
	// 这样可以得到最新的 N 条消息，且顺序正确
	subQuery := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit)

	err := r.db.WithContext(ctx).
		Table("(?) as t", subQuery).
		Order("created_at ASC").
		Find(&messages).Error

	return messages, err
}

// CountBySessionID 统计会话的消息数量
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - int64: 消息数量
//   - error: 数据库错误
func (r *MessageRepository) CountBySessionID(ctx context.Context, sessionID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Message{}).Where("session_id = ?", sessionID).Count(&count).Error
	return count, err
}

// DeleteBySessionID 删除会话的所有消息
// 通常在删除会话时使用（如果没有设置级联删除）
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - error: 数据库错误
func (r *MessageRepository) DeleteBySessionID(ctx context.Context, sessionID int64) error {
	return r.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&model.Message{}).Error
}

// GetLastUserMessage 获取会话的最后一条用户消息
// 用于重试功能
// 参数:
//   - ctx: 上下文
//   - sessionID: 会话ID
//
// 返回:
//   - *model.Message: 最后一条用户消息，没有则返回 nil
//   - error: 数据库错误
func (r *MessageRepository) GetLastUserMessage(ctx context.Context, sessionID int64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND role = ?", sessionID, model.MessageRoleUser).
		Order("created_at DESC").
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}
