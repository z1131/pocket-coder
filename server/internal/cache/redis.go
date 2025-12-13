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

// ==================== 在线状态管理 ====================
// 使用 Redis Set 存储在线设备列表，支持快速查询

// SetDesktopOnline 设置设备在线
// 当电脑端 WebSocket 连接成功时调用
// 参数:
//   - ctx: 上下文
//   - desktopID: 设备ID
//   - userID: 用户ID
//   - processID: 进程ID（用于区分重启）
//
// 返回:
//   - error: Redis 操作错误
func (c *RedisCache) SetDesktopOnline(ctx context.Context, desktopID, userID int64, processID string) error {
	pipe := c.client.Pipeline()

	// 添加到全局在线设备集合
	// SADD 如果元素已存在，不会重复添加
	pipe.SAdd(ctx, "online:desktops", desktopID)

	// 添加到用户的在线设备集合
	pipe.SAdd(ctx, fmt.Sprintf("user:%d:online_desktops", userID), desktopID)

	// 设置心跳时间，2分钟过期
	// 如果 2 分钟内没有更新心跳，Key 会自动删除
	pipe.Set(ctx, fmt.Sprintf("desktop:%d:heartbeat", desktopID), time.Now().Unix(), 2*time.Minute)

	// 存储 ProcessID
	pipe.Set(ctx, fmt.Sprintf("desktop:%d:pid", desktopID), processID, 0)

	_, err := pipe.Exec(ctx)
	return err
}

// GetDesktopProcessID 获取设备当前的 ProcessID
func (c *RedisCache) GetDesktopProcessID(ctx context.Context, desktopID int64) (string, error) {
	pid, err := c.client.Get(ctx, fmt.Sprintf("desktop:%d:pid", desktopID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return pid, err
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

// AppendTerminalHistory 追加终端历史记录
func (c *RedisCache) AppendTerminalHistory(ctx context.Context, sessionID int64, data []byte) error {
	key := fmt.Sprintf("session:history:%d", sessionID)
	// 使用 Append 命令
	if err := c.client.Append(ctx, key, string(data)).Err(); err != nil {
		return err
	}
	// 设置过期时间（例如 7 天）
	return c.client.Expire(ctx, key, 7*24*time.Hour).Err()
}

// GetTerminalHistory 获取终端历史记录
func (c *RedisCache) GetTerminalHistory(ctx context.Context, sessionID int64) ([]byte, error) {
	key := fmt.Sprintf("session:history:%d", sessionID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// ClearTerminalHistory 清除终端历史记录
func (c *RedisCache) ClearTerminalHistory(ctx context.Context, sessionID int64) error {
	key := fmt.Sprintf("session:history:%d", sessionID)
	return c.client.Del(ctx, key).Err()
}

// GetTerminalHistoryTail 获取终端历史记录的最后一部分（用于预览）
func (c *RedisCache) GetTerminalHistoryTail(ctx context.Context, sessionID int64, size int64) ([]byte, error) {
	key := fmt.Sprintf("session:history:%d", sessionID)
	// GETRANGE key start end
	// start 为负数表示倒数
	data, err := c.client.GetRange(ctx, key, -size, -1).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}
