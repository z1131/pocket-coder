// Package repository 提供数据访问层的实现
// 封装所有与数据库的交互操作
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"pocket-coder-server/internal/model"
)

// UserRepository 用户数据访问层
// 负责用户相关的所有数据库操作
type UserRepository struct {
	db *gorm.DB // GORM 数据库连接实例
}

// NewUserRepository 创建 UserRepository 实例
// 参数:
//   - db: GORM 数据库连接
//
// 返回:
//   - *UserRepository: 用户仓库实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建新用户
// 参数:
//   - ctx: 上下文，用于控制请求生命周期
//   - user: 用户对象，ID 字段会被自动填充
//
// 返回:
//   - error: 如果用户名或邮箱重复，会返回错误
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	// 使用 WithContext 确保数据库操作可以被取消
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据 ID 获取用户
// 参数:
//   - ctx: 上下文
//   - id: 用户ID
//
// 返回:
//   - *model.User: 用户对象，如果未找到返回 nil
//   - error: 数据库错误（不包括记录未找到）
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	// First 方法会按主键查询第一条记录
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		// 检查是否是"记录未找到"错误
		// 这是 GORM 特有的错误类型
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回 nil，不当作错误
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
// 用于登录验证
// 参数:
//   - ctx: 上下文
//   - username: 用户名
//
// 返回:
//   - *model.User: 用户对象，如果未找到返回 nil
//   - error: 数据库错误
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	// Where 方法添加查询条件
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
// 用于注册检查和密码找回
// 参数:
//   - ctx: 上下文
//   - email: 邮箱地址
//
// 返回:
//   - *model.User: 用户对象，如果未找到返回 nil
//   - error: 数据库错误
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
// 只更新非零值字段（GORM 的默认行为）
// 参数:
//   - ctx: 上下文
//   - user: 包含要更新字段的用户对象，必须包含 ID
//
// 返回:
//   - error: 数据库错误
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	// Save 方法会更新所有字段
	// 如果只想更新特定字段，使用 Updates 方法
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields 更新用户的指定字段
// 参数:
//   - ctx: 上下文
//   - id: 用户ID
//   - fields: 要更新的字段映射，如 map[string]interface{}{"avatar": "xxx"}
//
// 返回:
//   - error: 数据库错误
func (r *UserRepository) UpdateFields(ctx context.Context, id int64, fields map[string]interface{}) error {
	// Model 指定要操作的模型
	// Updates 更新指定的字段
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除用户
// 注意: 这是硬删除，如果需要软删除，请在模型中添加 gorm.DeletedAt 字段
// 参数:
//   - ctx: 上下文
//   - id: 用户ID
//
// 返回:
//   - error: 数据库错误
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// ExistsByUsername 检查用户名是否已存在
// 参数:
//   - ctx: 上下文
//   - username: 用户名
//
// 返回:
//   - bool: 是否存在
//   - error: 数据库错误
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否已存在
// 参数:
//   - ctx: 上下文
//   - email: 邮箱
//
// 返回:
//   - bool: 是否存在
//   - error: 数据库错误
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
