# 数据访问层 (Repository)

## 模块状态

- **状态**: ✅ 已完成
- **创建时间**: 2024-01-15
- **最后更新**: 2024-01-15

## 功能说明

Repository 层负责与数据库交互，封装所有 SQL 操作。类似于 Java 中的 Mapper/DAO 层。

## 设计原则

1. **单一职责**: 每个 Repository 只负责一个模型的数据操作
2. **接口抽象**: 定义接口便于测试和替换实现
3. **错误处理**: 统一的错误处理方式

## Repository 列表

### UserRepository

| 方法 | 说明 |
|------|------|
| Create(user) | 创建用户 |
| GetByID(id) | 根据 ID 获取用户 |
| GetByUsername(username) | 根据用户名获取用户 |
| GetByEmail(email) | 根据邮箱获取用户 |
| Update(user) | 更新用户信息 |
| Delete(id) | 删除用户 |

### DesktopRepository

| 方法 | 说明 |
|------|------|
| Create(desktop) | 创建设备 |
| GetByID(id) | 根据 ID 获取设备 |
| GetByUserID(userID) | 获取用户的所有设备 |
| GetByDeviceToken(token) | 根据设备令牌获取设备 |
| Update(desktop) | 更新设备信息 |
| UpdateStatus(id, status) | 更新设备状态 |
| Delete(id) | 删除设备 |

### SessionRepository

| 方法 | 说明 |
|------|------|
| Create(session) | 创建会话 |
| GetByID(id) | 根据 ID 获取会话 |
| GetByDesktopID(desktopID) | 获取设备的所有会话 |
| GetActiveByDesktopID(desktopID) | 获取设备的活跃会话 |
| Update(session) | 更新会话 |
| EndSession(id) | 结束会话 |
| Delete(id) | 删除会话 |

### MessageRepository

| 方法 | 说明 |
|------|------|
| Create(message) | 创建消息 |
| GetBySessionID(sessionID) | 获取会话的所有消息 |
| GetBySessionIDWithPagination(sessionID, page, size) | 分页获取消息 |

## 文件路径

- `internal/repository/user_repo.go`
- `internal/repository/desktop_repo.go`
- `internal/repository/session_repo.go`
- `internal/repository/message_repo.go`

## 使用示例

```go
// 创建 repository
userRepo := repository.NewUserRepository(db)

// 创建用户
user := &model.User{
    Username: "zhangsan",
    PasswordHash: "xxx",
}
err := userRepo.Create(ctx, user)

// 查询用户
user, err := userRepo.GetByUsername(ctx, "zhangsan")
```

## GORM 常用操作

```go
// 创建
db.Create(&user)

// 查询单条
db.First(&user, id)
db.Where("username = ?", name).First(&user)

// 查询多条
db.Where("user_id = ?", userID).Find(&desktops)

// 更新
db.Model(&user).Updates(map[string]interface{}{"name": "new_name"})

// 删除
db.Delete(&user, id)

// 预加载关联
db.Preload("Messages").Find(&session)
```

## 错误处理

```go
// 检查记录是否存在
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil // 返回 nil 表示未找到
}
```
