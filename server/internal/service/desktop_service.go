// Package service 提供业务逻辑层的实现
package service

import (
	"context"
	"errors"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
	"pocket-coder-server/pkg/util"
)

// 设备服务相关错误
var (
	ErrDesktopNotFound   = errors.New("设备不存在")
	ErrDesktopOffline    = errors.New("设备已离线")
	ErrNoPermission      = errors.New("无权限操作此设备")
)

// DesktopService 设备服务
// 处理电脑设备的管理操作
type DesktopService struct {
	desktopRepo *repository.DesktopRepository // 设备数据访问层
	sessionRepo *repository.SessionRepository // 会话数据访问层
	cache       *cache.RedisCache             // Redis 缓存
}

// NewDesktopService 创建 DesktopService 实例
func NewDesktopService(
	desktopRepo *repository.DesktopRepository,
	sessionRepo *repository.SessionRepository,
	cache *cache.RedisCache,
) *DesktopService {
	return &DesktopService{
		desktopRepo: desktopRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
	}
}

// DesktopResponse 设备响应（包含实时状态）
type DesktopResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	IP            *string `json:"ip,omitempty"`
	Status        string  `json:"status"` // 实时状态（从 Redis 获取）
	OSInfo        *string `json:"os_info,omitempty"`
	LastHeartbeat *string `json:"last_heartbeat,omitempty"`
}

// RegisterDesktopRequest 注册设备请求
type RegisterDesktopRequest struct {
	Name       string  `json:"name" binding:"required"`
	DeviceUUID string  `json:"device_uuid" binding:"required"` // 客户端持久化的设备 UUID
	IP         *string `json:"ip,omitempty"`                   // 设备 IP
	OSInfo     *string `json:"os_info,omitempty"`
}

// RegisterDesktopResult 注册设备结果
type RegisterDesktopResult struct {
	Desktop *DesktopResponse
	// DeviceToken 用于生成桌面专用 JWT 的设备令牌
	DeviceToken string
}

// ListDesktops 获取用户的设备列表
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - []DesktopResponse: 设备列表（包含实时状态）
//   - error: 操作错误
func (s *DesktopService) ListDesktops(ctx context.Context, userID int64) ([]DesktopResponse, error) {
	// 1. 从数据库获取设备列表
	desktops, err := s.desktopRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 从 Redis 获取在线设备列表
	onlineIDs, err := s.cache.GetUserOnlineDesktops(ctx, userID)
	if err != nil {
		// 如果 Redis 出错，降级为使用数据库中的状态
		onlineIDs = []int64{}
	}

	// 3. 将在线 ID 转换为 map 便于查找
	onlineMap := make(map[int64]bool)
	for _, id := range onlineIDs {
		onlineMap[id] = true
	}

	// 4. 构建响应
	result := make([]DesktopResponse, len(desktops))
	for i, desktop := range desktops {
		// 确定设备状态
		status := model.DesktopStatusOffline
		if onlineMap[desktop.ID] {
			status = model.DesktopStatusOnline
		}

		result[i] = DesktopResponse{
			ID:     desktop.ID,
			Name:   desktop.Name,
			Type:   desktop.Type,
			IP:     desktop.IP,
			Status: status,
			OSInfo: desktop.OSInfo,
		}

		// 格式化最后心跳时间
		if desktop.LastHeartbeat != nil {
			formatted := desktop.LastHeartbeat.Format("2006-01-02T15:04:05Z07:00")
			result[i].LastHeartbeat = &formatted
		}
	}

	return result, nil
}

// RegisterDesktop 为当前用户注册一台桌面设备
// 使用 device_uuid 作为设备唯一标识进行去重
// 如果设备已存在（按 device_uuid），则更新设备信息并返回
// 创建记录并返回基础信息与设备令牌（用于后续生成桌面 JWT）
func (s *DesktopService) RegisterDesktop(ctx context.Context, userID int64, req *RegisterDesktopRequest) (*RegisterDesktopResult, error) {
	if req == nil || req.Name == "" {
		return nil, errors.New("设备名称不能为空")
	}
	if req.DeviceUUID == "" {
		return nil, errors.New("设备标识不能为空")
	}

	// 按 device_uuid 查找设备（比 name 更可靠，即使改主机名也能识别同一设备）
	existing, err := s.desktopRepo.GetByUserIDAndDeviceUUID(ctx, userID, req.DeviceUUID)
	if err != nil {
		return nil, err
	}

	var desktop *model.Desktop
	var deviceToken string

	if existing != nil {
		// 设备已存在，更新信息（包括可能变更的主机名）
		deviceToken = existing.DeviceToken
		existing.Name = req.Name // 更新主机名（可能已变更）
		existing.IP = req.IP
		existing.OSInfo = req.OSInfo

		if err := s.desktopRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		desktop = existing
	} else {
		// 创建新设备
		deviceToken = util.GenerateDeviceToken()
		desktop = &model.Desktop{
			UserID:      userID,
			Name:        req.Name,
			DeviceUUID:  req.DeviceUUID,
			DeviceToken: deviceToken,
			Type:        model.DesktopTypeLocal,
			IP:          req.IP,
			OSInfo:      req.OSInfo,
			Status:      model.DesktopStatusOffline,
		}

		if err := s.desktopRepo.Create(ctx, desktop); err != nil {
			return nil, err
		}
	}

	resp := &DesktopResponse{
		ID:     desktop.ID,
		Name:   desktop.Name,
		Type:   desktop.Type,
		IP:     desktop.IP,
		Status: model.DesktopStatusOffline,
		OSInfo: desktop.OSInfo,
	}

	return &RegisterDesktopResult{
		Desktop:     resp,
		DeviceToken: deviceToken,
	}, nil
}

// GetDesktop 获取设备详情
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//
// 返回:
//   - *DesktopResponse: 设备详情
//   - error: 设备不存在或无权限返回错误
func (s *DesktopService) GetDesktop(ctx context.Context, userID, desktopID int64) (*DesktopResponse, error) {
	// 1. 获取设备
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

	// 3. 从 Redis 获取实时状态
	isOnline := s.cache.IsDesktopOnline(ctx, desktopID)
	status := model.DesktopStatusOffline
	if isOnline {
		status = model.DesktopStatusOnline
	}

	result := &DesktopResponse{
		ID:     desktop.ID,
		Name:   desktop.Name,
		Type:   desktop.Type,
		IP:     desktop.IP,
		Status: status,
		OSInfo: desktop.OSInfo,
	}

	if desktop.LastHeartbeat != nil {
		formatted := desktop.LastHeartbeat.Format("2006-01-02T15:04:05Z07:00")
		result.LastHeartbeat = &formatted
	}

	return result, nil
}

// ValidateDesktopOwnership 校验桌面归属与设备令牌
func (s *DesktopService) ValidateDesktopOwnership(ctx context.Context, desktopID int64, deviceToken string, userID int64) (bool, error) {
	desktop, err := s.desktopRepo.GetByID(ctx, desktopID)
	if err != nil {
		return false, err
	}
	if desktop == nil {
		return false, ErrDesktopNotFound
	}
	if desktop.UserID != userID {
		return false, ErrNoPermission
	}
	if desktop.DeviceToken != deviceToken {
		return false, ErrNoPermission
	}
	return true, nil
}

// UpdateDesktopRequest 更新设备请求
type UpdateDesktopRequest struct {
	Name *string `json:"name"` // 设备名称
}

// UpdateDesktop 更新设备信息
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//   - req: 更新请求
//
// 返回:
//   - *DesktopResponse: 更新后的设备信息
//   - error: 设备不存在或无权限返回错误
func (s *DesktopService) UpdateDesktop(ctx context.Context, userID, desktopID int64, req *UpdateDesktopRequest) (*DesktopResponse, error) {
	// 1. 获取设备
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

	// 3. 准备要更新的字段
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}

	// 4. 如果没有要更新的字段，直接返回当前设备信息
	if len(fields) == 0 {
		return s.GetDesktop(ctx, userID, desktopID)
	}

	// 5. 更新数据库
	if err := s.desktopRepo.UpdateFields(ctx, desktopID, fields); err != nil {
		return nil, err
	}

	// 6. 返回更新后的设备信息
	return s.GetDesktop(ctx, userID, desktopID)
}

// DeleteDesktop 删除设备
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID（用于权限验证）
//   - desktopID: 设备ID
//
// 返回:
//   - error: 设备不存在或无权限返回错误
func (s *DesktopService) DeleteDesktop(ctx context.Context, userID, desktopID int64) error {
	// 1. 获取设备
	desktop, err := s.desktopRepo.GetByID(ctx, desktopID)
	if err != nil {
		return err
	}
	if desktop == nil {
		return ErrDesktopNotFound
	}

	// 2. 验证权限
	if desktop.UserID != userID {
		return ErrNoPermission
	}

	// 3. 如果设备在线，先设为离线
	if s.cache.IsDesktopOnline(ctx, desktopID) {
		_ = s.cache.SetDesktopOffline(ctx, desktopID, userID)
	}

	// 4. 删除设备（级联删除关联的会话和消息）
	return s.desktopRepo.Delete(ctx, desktopID)
}

// SetDesktopOnline 设置设备在线
// 当电脑端 WebSocket 连接成功时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - userID: 用户ID
//   - processID: 进程ID（用于区分重启）
//
// 返回:
//   - error: 操作错误
func (s *DesktopService) SetDesktopOnline(ctx context.Context, desktopID, userID int64, processID string) error {
	// 1. 更新 Redis 在线状态
	if err := s.cache.SetDesktopOnline(ctx, desktopID, userID, processID); err != nil {
		return err
	}

	// 2. 更新数据库状态
	return s.desktopRepo.UpdateStatus(ctx, desktopID, model.DesktopStatusOnline)
}

// SetDesktopOffline 设置设备离线
// 当电脑端 WebSocket 断开时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - userID: 用户ID
//
// 返回:
//   - error: 操作错误
func (s *DesktopService) SetDesktopOffline(ctx context.Context, desktopID, userID int64) error {
	// 1. 更新 Redis 离线状态
	if err := s.cache.SetDesktopOffline(ctx, desktopID, userID); err != nil {
		return err
	}

	// 2. 更新数据库状态
	if err := s.desktopRepo.UpdateStatus(ctx, desktopID, model.DesktopStatusOffline); err != nil {
		return err
	}

	// 3. 结束设备上的所有活跃会话
	return s.sessionRepo.EndAllActiveByDesktopID(ctx, desktopID)
}

// UpdateHeartbeat 更新设备心跳
// 电脑端每隔一段时间调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - error: 操作错误
func (s *DesktopService) UpdateHeartbeat(ctx context.Context, desktopID int64) error {
	// 1. 更新 Redis 心跳
	if err := s.cache.UpdateHeartbeat(ctx, desktopID); err != nil {
		return err
	}

	// 2. 更新数据库心跳时间
	return s.desktopRepo.UpdateHeartbeat(ctx, desktopID)
}

// IsDesktopOnline 检查设备是否在线
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - bool: 是否在线
func (s *DesktopService) IsDesktopOnline(ctx context.Context, desktopID int64) bool {
	return s.cache.IsDesktopOnline(ctx, desktopID)
}

// GetDesktopByID 获取设备（不验证权限，内部使用）
func (s *DesktopService) GetDesktopByID(ctx context.Context, desktopID int64) (*model.Desktop, error) {
	return s.desktopRepo.GetByID(ctx, desktopID)
}
