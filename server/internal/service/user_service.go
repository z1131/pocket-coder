// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
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

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Email  *string `json:"email"`  // 邮箱
	Avatar *string `json:"avatar"` // 头像 URL
}

// UpdateProfile 更新用户资料
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//   - req: 更新请求
//
// 返回:
//   - *model.User: 更新后的用户信息
//   - error: 操作错误
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *UpdateProfileRequest) (*model.User, error) {
	// 1. 获取当前用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 2. 准备要更新的字段
	fields := make(map[string]interface{})

	// 如果提供了邮箱且不同于当前邮箱
	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		if *req.Email != "" {
			existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email)
			if err != nil {
				return nil, err
			}
			if existingUser != nil && existingUser.ID != userID {
				return nil, ErrEmailExists
			}
		}
		fields["email"] = req.Email
	}

	// 如果提供了头像 URL
	if req.Avatar != nil {
		fields["avatar"] = req.Avatar
	}

	// 3. 如果没有要更新的字段，直接返回
	if len(fields) == 0 {
		return user, nil
	}

	// 4. 更新数据库
	if err := s.userRepo.UpdateFields(ctx, userID, fields); err != nil {
		return nil, err
	}

	// 5. 重新获取更新后的用户信息
	return s.userRepo.GetByID(ctx, userID)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码
}

// ChangePassword 修改密码
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//   - req: 修改密码请求
//
// 返回:
//   - error: 旧密码错误等情况返回错误
func (s *UserService) ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error {
	// 1. 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. 验证旧密码
	if !checkPassword(req.OldPassword, user.PasswordHash) {
		return ErrPasswordWrong
	}

	// 3. 对新密码进行哈希
	newHash, err := hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// 4. 更新密码
	return s.userRepo.UpdateFields(ctx, userID, map[string]interface{}{
		"password_hash": newHash,
	})
}

// 密码工具函数
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
