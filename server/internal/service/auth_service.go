// Package service 提供业务逻辑层的实现
// 服务层封装具体的业务逻辑，协调 Repository 和 Cache
package service

import (
	"context"
	"errors"
	"time"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
	"pocket-coder-server/pkg/jwt"
	"pocket-coder-server/pkg/util"
)

// 定义业务错误
var (
	ErrUserExists       = errors.New("用户名已存在")
	ErrEmailExists      = errors.New("邮箱已被注册")
	ErrUserNotFound     = errors.New("用户不存在")
	ErrPasswordWrong    = errors.New("密码错误")
	ErrDeviceCodeNotFound = errors.New("设备授权码不存在或已过期")
	ErrDeviceCodePending  = errors.New("设备授权码等待授权中")
	ErrInvalidUserCode    = errors.New("无效的用户码")
)

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
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// Register 用户注册
// 参数:
//   - ctx: 上下文
//   - req: 注册请求
//
// 返回:
//   - *RegisterResponse: 注册成功返回用户信息
//   - error: 注册失败返回错误（用户名/邮箱已存在等）
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 1. 检查用户名是否已存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 2. 如果提供了邮箱，检查邮箱是否已存在
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailExists
		}
	}

	// 3. 对密码进行哈希
	// 使用 bcrypt 算法，自动添加盐值
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 4. 创建用户
	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Status:       1, // 正常状态
	}

	// 如果提供了邮箱，设置邮箱字段
	if req.Email != "" {
		user.Email = &req.Email
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
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
	// 1. 根据用户名查找用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
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

// DeviceCodeRequest 设备码请求
type DeviceCodeRequest struct {
	DeviceToken string `json:"device_token"` // 设备令牌，首次为空则生成
	DeviceName  string `json:"device_name"`  // 设备名称
	OSInfo      string `json:"os_info"`      // 操作系统信息
}

// DeviceCodeResponse 设备码响应
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`      // 设备码（内部使用）
	UserCode        string `json:"user_code"`        // 用户码（用户输入）
	VerificationURI string `json:"verification_uri"` // 验证 URI
	ExpiresIn       int    `json:"expires_in"`       // 过期时间（秒）
	Interval        int    `json:"interval"`         // 轮询间隔（秒）
}

// RequestDeviceCode 请求设备授权码
// 电脑端调用此接口获取设备码
// 参数:
//   - ctx: 上下文
//   - req: 设备码请求
//
// 返回:
//   - *DeviceCodeResponse: 设备码信息
//   - error: 操作错误
func (s *AuthService) RequestDeviceCode(ctx context.Context, req *DeviceCodeRequest) (*DeviceCodeResponse, error) {
	// 1. 如果没有设备令牌，生成一个
	deviceToken := req.DeviceToken
	if deviceToken == "" {
		deviceToken = util.GenerateDeviceToken()
	}

	// 2. 生成设备码和用户码
	deviceCode := util.GenerateDeviceCode()
	userCode := util.GenerateUserCode()

	// 3. 创建设备码信息
	info := &cache.DeviceCodeInfo{
		DeviceToken: deviceToken,
		UserCode:    userCode,
		Status:      "pending", // 等待授权
		DeviceName:  req.DeviceName,
		OSInfo:      req.OSInfo,
	}

	// 4. 存入 Redis，15 分钟过期
	if err := s.cache.CreateDeviceCode(ctx, deviceCode, info); err != nil {
		return nil, err
	}

	return &DeviceCodeResponse{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: "https://pocket-coder.com/device", // TODO: 从配置读取
		ExpiresIn:       900,                                // 15 分钟
		Interval:        5,                                  // 5 秒轮询
	}, nil
}

// DeviceStatusResponse 设备授权状态响应
type DeviceStatusResponse struct {
	Status      string `json:"status"`                 // pending: 等待, authorized: 已授权
	AccessToken string `json:"access_token,omitempty"` // 授权成功时返回
	DesktopID   int64  `json:"desktop_id,omitempty"`   // 授权成功时返回
}

// GetDeviceStatus 获取设备授权状态
// 电脑端轮询调用此接口
// 参数:
//   - ctx: 上下文
//   - deviceCode: 设备码
//
// 返回:
//   - *DeviceStatusResponse: 授权状态
//   - error: 设备码不存在或已过期返回错误
func (s *AuthService) GetDeviceStatus(ctx context.Context, deviceCode string) (*DeviceStatusResponse, error) {
	// 1. 从 Redis 获取设备码信息
	info, err := s.cache.GetDeviceCode(ctx, deviceCode)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, ErrDeviceCodeNotFound
	}

	// 2. 如果还在等待授权
	if info.Status == "pending" {
		return &DeviceStatusResponse{
			Status: "pending",
		}, nil
	}

	// 3. 已授权，创建或更新设备记录
	// 检查设备是否已存在
	desktop, err := s.desktopRepo.GetByDeviceToken(ctx, info.DeviceToken)
	if err != nil {
		return nil, err
	}

	if desktop == nil {
		// 创建新设备
		desktop = &model.Desktop{
			UserID:      info.UserID,
			Name:        info.DeviceName,
			DeviceToken: info.DeviceToken,
			Type:        model.DesktopTypeLocal,
			AgentType:   "claude-code",
			Status:      model.DesktopStatusOffline,
		}
		if info.OSInfo != "" {
			desktop.OSInfo = &info.OSInfo
		}
		if err := s.desktopRepo.Create(ctx, desktop); err != nil {
			return nil, err
		}
	} else {
		// 更新设备的用户ID（重新授权给新用户）
		desktop.UserID = info.UserID
		if info.DeviceName != "" {
			desktop.Name = info.DeviceName
		}
		if info.OSInfo != "" {
			desktop.OSInfo = &info.OSInfo
		}
		if err := s.desktopRepo.Update(ctx, desktop); err != nil {
			return nil, err
		}
	}

	// 4. 生成设备 Token
	accessToken, err := s.jwtService.GenerateDesktopToken(info.UserID, desktop.ID, info.DeviceToken)
	if err != nil {
		return nil, err
	}

	// 5. 清理 Redis 中的设备码
	_ = s.cache.DeleteDeviceCode(ctx, deviceCode, info.UserCode)

	return &DeviceStatusResponse{
		Status:      "authorized",
		AccessToken: accessToken,
		DesktopID:   desktop.ID,
	}, nil
}

// AuthorizeDevice 授权设备
// 手机端用户输入用户码后调用
// 参数:
//   - ctx: 上下文
//   - userID: 授权用户ID
//   - userCode: 用户码
//
// 返回:
//   - error: 用户码无效返回错误
func (s *AuthService) AuthorizeDevice(ctx context.Context, userID int64, userCode string) error {
	// 1. 通过用户码获取设备码
	deviceCode, err := s.cache.GetDeviceCodeByUserCode(ctx, userCode)
	if err != nil {
		return err
	}
	if deviceCode == "" {
		return ErrInvalidUserCode
	}

	// 2. 授权设备码
	return s.cache.AuthorizeDeviceCode(ctx, deviceCode, userID)
}
