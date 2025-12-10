// Package cache 提供 Redis 缓存操作的封装
// 处理设备授权码、在线状态、JWT 黑名单等需要快速访问的数据
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"pocket-coder-server/internal/config"
)

// RedisCache 封装 Redis 客户端，提供业务相关的缓存操作
type RedisCache struct {
	client *redis.Client // Redis 客户端实例
}

// NewRedisCache 创建 RedisCache 实例
// 参数:
//   - cfg: 应用配置（包含 Redis 连接信息）
//
// 返回:
//   - *RedisCache: 缓存实例
//   - error: 连接错误
func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username, // 阿里云 Redis 需要用户名
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Close 关闭 Redis 连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// ==================== 设备授权码相关 ====================
// 用于实现 Device Flow 登录流程
// 电脑端获取设备码 -> 用户在手机端输入用户码授权 -> 电脑端轮询获取授权结果

// DeviceCodeInfo 设备授权码信息
// 存储在 Redis 中，用于设备登录流程
type DeviceCodeInfo struct {
	DeviceToken string `json:"device_token"` // 设备唯一标识
	UserCode    string `json:"user_code"`    // 用户输入的短码，如 "ABCD-1234"
	Status      string `json:"status"`       // pending: 等待授权, authorized: 已授权
	UserID      int64  `json:"user_id,omitempty"` // 授权用户ID，授权后填充
	DeviceName  string `json:"device_name,omitempty"` // 设备名称
	OSInfo      string `json:"os_info,omitempty"`     // 操作系统信息
}

// CreateDeviceCode 创建设备授权码
// 将授权码信息存入 Redis，设置 15 分钟过期
// 参数:
//   - ctx: 上下文
//   - code: 设备码（长码，内部使用）
//   - info: 授权码信息
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) CreateDeviceCode(ctx context.Context, code string, info *DeviceCodeInfo) error {
	// 将结构体序列化为 JSON
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal device code info: %w", err)
	}

	// 使用 Pipeline 一次性执行多个命令
	// Pipeline 可以减少网络往返次数，提高性能
	pipe := c.client.Pipeline()

	// 存储设备码信息
	// Key: device_code:{code}
	// TTL: 15分钟
	pipe.Set(ctx, fmt.Sprintf("device_code:%s", code), data, 15*time.Minute)

	// 存储用户码到设备码的映射（反向索引）
	// 用户输入用户码时，可以快速找到对应的设备码
	// Key: user_code:{user_code}
	pipe.Set(ctx, fmt.Sprintf("user_code:%s", info.UserCode), code, 15*time.Minute)

	// 执行 Pipeline 中的所有命令
	_, err = pipe.Exec(ctx)
	return err
}

// GetDeviceCode 获取设备授权码信息
// 参数:
//   - ctx: 上下文
//   - code: 设备码
//
// 返回:
//   - *DeviceCodeInfo: 授权码信息，不存在或已过期返回 nil
//   - error: Redis 操作错误
func (c *RedisCache) GetDeviceCode(ctx context.Context, code string) (*DeviceCodeInfo, error) {
	key := fmt.Sprintf("device_code:%s", code)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil 表示 Key 不存在
		if err == redis.Nil {
			return nil, nil // 返回 nil 表示不存在或已过期
		}
		return nil, err
	}

	var info DeviceCodeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device code info: %w", err)
	}
	return &info, nil
}

// GetDeviceCodeByUserCode 通过用户码获取设备码
// 用户在手机端输入用户码时调用
// 参数:
//   - ctx: 上下文
//   - userCode: 用户码（如 "ABCD-1234"）
//
// 返回:
//   - string: 设备码，不存在返回空字符串
//   - error: Redis 操作错误
func (c *RedisCache) GetDeviceCodeByUserCode(ctx context.Context, userCode string) (string, error) {
	deviceCode, err := c.client.Get(ctx, fmt.Sprintf("user_code:%s", userCode)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return deviceCode, err
}

// AuthorizeDeviceCode 授权设备码
// 用户在手机端确认授权后调用
// 参数:
//   - ctx: 上下文
//   - code: 设备码
//   - userID: 授权用户ID
//
// 返回:
//   - error: 如果设备码不存在返回错误
func (c *RedisCache) AuthorizeDeviceCode(ctx context.Context, code string, userID int64) error {
	// 获取当前设备码信息
	info, err := c.GetDeviceCode(ctx, code)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("device code not found or expired")
	}

	// 更新状态和用户ID
	info.Status = "authorized"
	info.UserID = userID

	// 序列化更新后的信息
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// 获取原有 TTL，保持原有过期时间
	key := fmt.Sprintf("device_code:%s", code)
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute // 如果 TTL 异常，设置默认 5 分钟
	}

	// 更新 Redis 中的值
	return c.client.Set(ctx, key, data, ttl).Err()
}

// DeleteDeviceCode 删除设备授权码
// 授权完成后调用清理
// 参数:
//   - ctx: 上下文
//   - code: 设备码
//   - userCode: 用户码
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) DeleteDeviceCode(ctx context.Context, code, userCode string) error {
	pipe := c.client.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("device_code:%s", code))
	pipe.Del(ctx, fmt.Sprintf("user_code:%s", userCode))
	_, err := pipe.Exec(ctx)
	return err
}

// ==================== 在线状态管理 ====================
// 使用 Redis Set 存储在线设备列表，支持快速查询

// SetDesktopOnline 设置设备在线
// 当电脑端 WebSocket 连接成功时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - userID: 用户ID
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) SetDesktopOnline(ctx context.Context, desktopID, userID int64) error {
	pipe := c.client.Pipeline()

	// 添加到全局在线设备集合
	// SADD 如果元素已存在，不会重复添加
	pipe.SAdd(ctx, "online:desktops", desktopID)

	// 添加到用户的在线设备集合
	pipe.SAdd(ctx, fmt.Sprintf("user:%d:online_desktops", userID), desktopID)

	// 设置心跳时间，2分钟过期
	// 如果 2 分钟内没有更新心跳，Key 会自动删除
	pipe.Set(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID), time.Now().Unix(), 2*time.Minute)

	_, err := pipe.Exec(ctx)
	return err
}

// SetDesktopOffline 设置设备离线
// 当电脑端 WebSocket 断开时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - userID: 用户ID
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) SetDesktopOffline(ctx context.Context, desktopID, userID int64) error {
	pipe := c.client.Pipeline()

	// 从全局在线集合移除
	// SREM 如果元素不存在，不会报错
	pipe.SRem(ctx, "online:desktops", desktopID)

	// 从用户的在线设备集合移除
	pipe.SRem(ctx, fmt.Sprintf("user:%d:online_desktops", userID), desktopID)

	// 删除心跳 Key
	pipe.Del(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID))

	// 删除活跃会话记录
	pipe.Del(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID))

	_, err := pipe.Exec(ctx)
	return err
}

// UpdateHeartbeat 更新设备心跳
// 电脑端每 30 秒调用一次，刷新在线状态
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) UpdateHeartbeat(ctx context.Context, desktopID int64) error {
	// 设置心跳时间，2分钟过期
	// 如果电脑端正常发送心跳（每30秒），这个 Key 会一直存在
	// 如果电脑端断开（停止发送心跳），2分钟后 Key 会自动删除
	return c.client.Set(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID), time.Now().Unix(), 2*time.Minute).Err()
}

// IsDesktopOnline 检查设备是否在线
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - bool: 是否在线
func (c *RedisCache) IsDesktopOnline(ctx context.Context, desktopID int64) bool {
	// SISMEMBER 检查元素是否在集合中，O(1) 复杂度
	return c.client.SIsMember(ctx, "online:desktops", desktopID).Val()
}

// GetUserOnlineDesktops 获取用户的在线设备列表
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - []int64: 在线设备ID列表
//   - error: Redis 操作错误
func (c *RedisCache) GetUserOnlineDesktops(ctx context.Context, userID int64) ([]int64, error) {
	// SMEMBERS 获取集合的所有成员
	result, err := c.client.SMembers(ctx, fmt.Sprintf("user:%d:online_desktops", userID)).Result()
	if err != nil {
		return nil, err
	}

	// 将字符串转换为 int64
	ids := make([]int64, 0, len(result))
	for _, s := range result {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue // 跳过无效的值
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetAllOnlineDesktops 获取所有在线设备
// 用于管理后台或监控
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []int64: 所有在线设备ID
//   - error: Redis 操作错误
func (c *RedisCache) GetAllOnlineDesktops(ctx context.Context) ([]int64, error) {
	result, err := c.client.SMembers(ctx, "online:desktops").Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(result))
	for _, s := range result {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ==================== 会话缓存 ====================

// SetActiveSession 设置设备的当前活跃会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - sessionID: 会话ID
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) SetActiveSession(ctx context.Context, desktopID, sessionID int64) error {
	// 不设置过期时间，因为会话可能持续很长时间
	// 设备离线时会清理
	return c.client.Set(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID), sessionID, 0).Err()
}

// GetActiveSession 获取设备的当前活跃会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - int64: 会话ID，没有活跃会话返回 0
//   - error: Redis 操作错误
func (c *RedisCache) GetActiveSession(ctx context.Context, desktopID int64) (int64, error) {
	result, err := c.client.Get(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID)).Int64()
	if err == redis.Nil {
		return 0, nil // 没有活跃会话
	}
	return result, err
}

// ClearActiveSession 清除设备的活跃会话
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) ClearActiveSession(ctx context.Context, desktopID int64) error {
	return c.client.Del(ctx, fmt.Sprintf("desktop:%d:active_session", desktopID)).Err()
}

// ==================== JWT 黑名单 ====================
// 用于实现 Token 强制失效（登出）功能

// BlacklistToken 将 Token 加入黑名单
// 登出时调用，使当前 Token 失效
// 参数:
//   - ctx: 上下文
//   - tokenHash: Token 的哈希值（不存储原始 Token）
//   - expireAt: Token 的原始过期时间
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) BlacklistToken(ctx context.Context, tokenHash string, expireAt time.Time) error {
	// 计算剩余有效时间
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		// Token 已过期，无需加入黑名单
		return nil
	}

	// 设置黑名单 Key
	// 值为 "1" 表示已加入黑名单
	// TTL 设置为 Token 的剩余有效期，过期后自动删除（因为 Token 本身也过期了）
	return c.client.Set(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenHash), "1", ttl).Err()
}

// IsTokenBlacklisted 检查 Token 是否在黑名单中
// JWT 验证中间件调用
// 参数:
//   - ctx: 上下文
//   - tokenHash: Token 的哈希值
//
// 返回:
//   - bool: 是否在黑名单中
func (c *RedisCache) IsTokenBlacklisted(ctx context.Context, tokenHash string) bool {
	// EXISTS 命令返回存在的 Key 数量
	return c.client.Exists(ctx, fmt.Sprintf("jwt:blacklist:%s", tokenHash)).Val() > 0
}

// ==================== Pub/Sub ====================
// 用于多服务实例间的消息广播

// PublishUserMessage 发布用户消息
// 用于多实例部署时的跨实例通信
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//   - message: 消息内容（会被 JSON 序列化）
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) PublishUserMessage(ctx context.Context, userID int64, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	// PUBLISH 发布消息到指定频道
	// 所有订阅该频道的客户端都会收到消息
	return c.client.Publish(ctx, fmt.Sprintf("user:%d:messages", userID), data).Err()
}

// SubscribeUserMessages 订阅用户消息
// 返回 PubSub 对象，调用方负责关闭
// 参数:
//   - ctx: 上下文
//   - userID: 用户ID
//
// 返回:
//   - *redis.PubSub: PubSub 订阅对象
func (c *RedisCache) SubscribeUserMessages(ctx context.Context, userID int64) *redis.PubSub {
	return c.client.Subscribe(ctx, fmt.Sprintf("user:%d:messages", userID))
}

// PublishDesktopStatus 发布设备状态变更
// 用于通知其他服务实例设备状态变化
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - status: 新状态
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) PublishDesktopStatus(ctx context.Context, desktopID int64, status string) error {
	data, _ := json.Marshal(map[string]interface{}{
		"desktop_id": desktopID,
		"status":     status,
		"timestamp":  time.Now().Unix(),
	})
	return c.client.Publish(ctx, "desktop:status", data).Err()
}

// SubscribeDesktopStatus 订阅设备状态变更
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *redis.PubSub: PubSub 订阅对象
func (c *RedisCache) SubscribeDesktopStatus(ctx context.Context) *redis.PubSub {
	return c.client.Subscribe(ctx, "desktop:status")
}

// ==================== 通用方法 ====================

// Ping 检查 Redis 连接
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 如果连接失败返回错误
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
