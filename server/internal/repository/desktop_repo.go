// Package repository 提供数据访问层的实现
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"pocket-coder-server/internal/model"
)

// DesktopRepository 电脑设备数据访问层
// 负责设备相关的所有数据库操作
type DesktopRepository struct {
	db *gorm.DB
}

// NewDesktopRepository 创建 DesktopRepository 实例
func NewDesktopRepository(db *gorm.DB) *DesktopRepository {
	return &DesktopRepository{db: db}
}

// Create 创建新设备记录
// 参数:
//   - ctx: 上下文
//   - desktop: 设备对象，ID 和 CreatedAt 会被自动填充
//
// 返回:
//   - error: 如果 device_token 重复会返回错误
func (r *DesktopRepository) Create(ctx context.Context, desktop *model.Desktop) error {
	return r.db.WithContext(ctx).Create(desktop).Error
}

// GetByID 根据 ID 获取设备
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//
// 返回:
//   - *model.Desktop: 设备对象，未找到返回 nil
//   - error: 数据库错误
func (r *DesktopRepository) GetByID(ctx context.Context, id int64) (*model.Desktop, error) {
	var desktop model.Desktop
	err := r.db.WithContext(ctx).First(&desktop, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &desktop, nil
}

// GetByIDWithUser 根据 ID 获取设备，同时预加载用户信息
// 用于需要同时获取设备和所属用户的场景
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//
// 返回:
//   - *model.Desktop: 包含 User 字段的设备对象
//   - error: 数据库错误
func (r *DesktopRepository) GetByIDWithUser(ctx context.Context, id int64) (*model.Desktop, error) {
	var desktop model.Desktop
	// Preload 会自动加载关联的 User 对象
	err := r.db.WithContext(ctx).Preload("User").First(&desktop, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &desktop, nil
}

// GetByUserID 获取用户的所有设备
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - []model.Desktop: 设备列表，可能为空
//   - error: 数据库错误
func (r *DesktopRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Desktop, error) {
	var desktops []model.Desktop
	// Find 用于查询多条记录
	// Order 按创建时间倒序排列，最新的在前面
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&desktops).Error
	return desktops, err
}

// GetByDeviceToken 根据设备令牌获取设备
// 用于设备认证
// 参数:
//   - ctx: 上下文
//   - deviceToken: 设备唯一标识符
//
// 返回:
//   - *model.Desktop: 设备对象，未找到返回 nil
//   - error: 数据库错误
func (r *DesktopRepository) GetByDeviceToken(ctx context.Context, deviceToken string) (*model.Desktop, error) {
	var desktop model.Desktop
	err := r.db.WithContext(ctx).Where("device_token = ?", deviceToken).First(&desktop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &desktop, nil
}

// Update 更新设备信息
// 参数:
//   - ctx: 上下文
//   - desktop: 包含要更新字段的设备对象，必须包含 ID
//
// 返回:
//   - error: 数据库错误
func (r *DesktopRepository) Update(ctx context.Context, desktop *model.Desktop) error {
	return r.db.WithContext(ctx).Save(desktop).Error
}

// UpdateFields 更新设备的指定字段
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//   - fields: 要更新的字段映射
//
// 返回:
//   - error: 数据库错误
func (r *DesktopRepository) UpdateFields(ctx context.Context, id int64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Desktop{}).Where("id = ?", id).Updates(fields).Error
}

// UpdateStatus 更新设备状态
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//   - status: 新状态 (online/offline/busy)
//
// 返回:
//   - error: 数据库错误
func (r *DesktopRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.Desktop{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateHeartbeat 更新设备心跳时间
// 每次电脑端发送心跳时调用
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//
// 返回:
//   - error: 数据库错误
func (r *DesktopRepository) UpdateHeartbeat(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.Desktop{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"status":         model.DesktopStatusOnline,
		}).Error
}

// Delete 删除设备
// 注意: 由于外键约束设置了 ON DELETE CASCADE，
// 删除设备时会自动删除关联的会话和消息
// 参数:
//   - ctx: 上下文
//   - id: 设备ID
//
// 返回:
//   - error: 数据库错误
func (r *DesktopRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Desktop{}, id).Error
}

// ExistsByDeviceToken 检查设备令牌是否已存在
// 参数:
//   - ctx: 上下文
//   - deviceToken: 设备令牌
//
// 返回:
//   - bool: 是否存在
//   - error: 数据库错误
func (r *DesktopRepository) ExistsByDeviceToken(ctx context.Context, deviceToken string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Desktop{}).Where("device_token = ?", deviceToken).Count(&count).Error
	return count > 0, err
}

// CountByUserID 统计用户的设备数量
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - int64: 设备数量
//   - error: 数据库错误
func (r *DesktopRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Desktop{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetOnlineByUserID 获取用户的在线设备
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - []model.Desktop: 在线设备列表
//   - error: 数据库错误
func (r *DesktopRepository) GetOnlineByUserID(ctx context.Context, userID int64) ([]model.Desktop, error) {
	var desktops []model.Desktop
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.DesktopStatusOnline).
		Find(&desktops).Error
	return desktops, err
}
