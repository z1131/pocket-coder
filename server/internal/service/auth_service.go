// Package service 提供业务逻辑层的实现
// 服务层封装具体的业务逻辑，协调 Repository 和 Cache
package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
	"pocket-coder-server/pkg/jwt"
	"pocket-coder-server/pkg/util"
)

// 定义业务错误
var (
	ErrUserExists         = errors.New("用户名已存在")
	ErrEmailExists        = errors.New("邮箱已被注册")
	ErrPhoneExists        = errors.New("手机号已被注册")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrPasswordWrong      = errors.New("密码错误")
	ErrInvalidUsername    = errors.New("用户名只能包含字母、数字和下划线，长度3-20")
)

// 用户名验证正则：只允许字母、数字、下划线，长度3-20
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

// validateUsername 验证用户名格式
func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

// AuthService 认证服务
// 处理用户注册、登录、登出以及设备授权
type AuthService struct {
	userRepo    *repository.UserRepository    // 用户数据访问层
	desktopRepo *repository.DesktopRepository // 设备数据访问层
	cache       *cache.RedisCache             // Redis 缓存
	jwtService  *jwt.JWTService               // JWT 服务
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(
	userRepo *repository.UserRepository,
	desktopRepo *repository.DesktopRepository,
	cache *cache.RedisCache,
	jwtService *jwt.JWTService,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		desktopRepo: desktopRepo,
		cache:       cache,
		jwtService:  jwtService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名
	Password string `json:"password" binding:"required,min=6"`        // 密码
	Email    string `json:"email" binding:"omitempty,email"`          // 邮箱（可选）
	Phone    string `json:"phone" binding:"omitempty,e164"`           // 手机号（可选）
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Token *LoginResponse `json:"token"`
	User  *model.User    `json:"user"`
}

// Register 用户注册
// 参数:
//   - ctx: 上下文
//   - req: 注册请求
//
// 返回:
//   - *RegisterResponse: 注册成功返回 Token 和用户信息
//   - error: 注册失败返回错误
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. 验证用户名格式
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}

	// 2. 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 3. 如果提供了邮箱，检查邮箱是否已存在
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailExists
		}
	}

	// 4. 如果提供了手机号，检查是否已存在
	if req.Phone != "" {
		exists, err = s.userRepo.ExistsByPhone(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPhoneExists
		}
	}

	// 5. 对密码进行哈希
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 6. 创建用户
	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Status:       1, // 正常状态
	}

	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 7. 自动登录（生成 Token）
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	tokenResp := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessExpire().Seconds()),
		User:         user,
	}

	return &RegisterResponse{
		Token: tokenResp,
		User:  user,
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 用户名/邮箱/手机号
	Password   string `json:"password" binding:"required"`   // 密码
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`  // 访问令牌
	RefreshToken string      `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int64       `json:"expires_in"`    // 过期时间（秒）
	User         *model.User `json:"user"`          // 用户信息
}

// Login 用户登录
// 参数:
//   - ctx: 上下文
//   - req: 登录请求
//
// 返回:
//   - *LoginResponse: 登录成功返回 Token 和用户信息
//   - error: 登录失败返回错误（用户不存在/密码错误）
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 1. 根据 标识符(用户名/邮箱/手机号) 查找用户
	user, err := s.userRepo.GetByIdentifier(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 2. 验证密码
	if !util.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrPasswordWrong
	}

	// 3. 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 4. 生成 Access Token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 5. 生成 Refresh Token
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessExpire().Seconds()),
		User:         user,
	}, nil
}

// Logout 用户登出
// 将 Token 加入黑名单
// 参数:
//   - ctx: 上下文
//   - tokenHash: Token 的哈希值
//   - expireAt: Token 的过期时间
//
// 返回:
//   - error: 操作错误
func (s *AuthService) Logout(ctx context.Context, tokenHash string, expireAt time.Time) error {
	// 将 Token 加入 Redis 黑名单
	// TTL 设为 Token 的剩余有效期
	return s.cache.BlacklistToken(ctx, tokenHash, expireAt)
}

// RefreshTokenResponse 刷新 Token 响应
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"` // 新的访问令牌
	ExpiresIn   int64  `json:"expires_in"`   // 过期时间（秒）
}

// RefreshToken 刷新 Access Token
// 参数:
//   - ctx: 上下文
//   - refreshToken: Refresh Token
//
// 返回:
//   - *RefreshTokenResponse: 新的 Access Token
//   - error: 刷新失败返回错误
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	// 1. 验证 Refresh Token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 检查用户是否仍然存在且正常
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 3. 生成新的 Access Token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtService.GetAccessExpire().Seconds()),
	}, nil
}
