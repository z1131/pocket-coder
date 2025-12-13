// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"errors"

	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
)

// 用户服务相关错误
var (
	ErrUserDisabled = errors.New("用户已被禁用")
)

// UserService 用户服务
// 处理用户信息的查询和更新
type UserService struct {
	userRepo *repository.UserRepository // 用户数据访问层
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetProfile 获取用户资料
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - *model.User: 用户信息
//   - error: 用户不存在返回错误
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
