// Package main 是服务端的入口点
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"pocket-coder-server/internal/cache"
	"pocket-coder-server/internal/config"
	"pocket-coder-server/internal/handler"
	"pocket-coder-server/internal/middleware"
	"pocket-coder-server/internal/model"
	"pocket-coder-server/internal/repository"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/internal/websocket"
	"pocket-coder-server/pkg/jwt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./configs")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化 Redis
	redisCache, err := cache.NewRedisCache(cfg)
	if err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	// 初始化 JWT 服务
	jwtService := jwt.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
	)

	// 初始化 Repository 层
	userRepo := repository.NewUserRepository(db)
	desktopRepo := repository.NewDesktopRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// 初始化 Service 层
	authService := service.NewAuthService(userRepo, desktopRepo, redisCache, jwtService)
	userService := service.NewUserService(userRepo)
	desktopService := service.NewDesktopService(desktopRepo, sessionRepo, redisCache)
	sessionService := service.NewSessionService(sessionRepo, messageRepo, desktopRepo, redisCache)

	// 初始化 WebSocket Hub
	wsHub := websocket.NewHub(desktopService, sessionService, redisCache)
	go wsHub.Run() // 在单独的 goroutine 中运行

	// 初始化 Handler 层
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	desktopHandler := handler.NewDesktopHandler(desktopService, jwtService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	wsHandler := websocket.NewHandler(wsHub, desktopService, cfg.JWT.Secret)

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	router := gin.New()

	// 全局中间件
	router.Use(gin.Recovery())                    // 恢复 panic
	router.Use(middleware.LoggerMiddleware())     // 请求日志
	router.Use(middleware.CORSMiddleware())       // CORS

	// 注册路由
	registerRoutes(router, jwtService, redisCache, authHandler, userHandler, desktopHandler, sessionHandler, wsHandler)

	// 创建 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 创建关闭上下文，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// 关闭 Redis 连接
	if err := redisCache.Close(); err != nil {
		log.Printf("Failed to close redis: %v", err)
	}

	log.Println("Server exited")
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// 构建 DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
		cfg.MySQL.Charset,
	)

	// 配置 GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.Server.Mode == "release" {
		gormLogger = logger.Default.LogMode(logger.Warn)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MySQL.MaxLifetime) * time.Second)

	log.Println("Database connected successfully")
	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&model.User{},
		&model.Desktop{},
		&model.Session{},
		&model.Message{},
	); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	log.Println("Database migrations completed")
	return nil
}

// registerRoutes 注册所有路由
func registerRoutes(
	router *gin.Engine,
	jwtService *jwt.JWTService,
	redisCache *cache.RedisCache,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	desktopHandler *handler.DesktopHandler,
	sessionHandler *handler.SessionHandler,
	wsHandler *websocket.Handler,
) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")

	// 认证相关（无需登录）
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)              // 刷新 Token
		auth.POST("/logout", authHandler.Logout)                      // 登出
		auth.POST("/device/code", authHandler.RequestDeviceCode)
		auth.GET("/device/status", authHandler.GetDeviceStatus)
		auth.POST("/device/authorize", authHandler.AuthorizeDevice)
	}

	// 用户相关（需要登录）
	users := v1.Group("/users")
	users.Use(middleware.AuthMiddleware(jwtService, redisCache))
	{
		users.GET("/me", userHandler.GetProfile)
		users.PUT("/me", userHandler.UpdateProfile)
		users.PUT("/me/password", userHandler.ChangePassword)
	}

	// 设备相关（需要登录）
	desktops := v1.Group("/desktops")
	desktops.Use(middleware.AuthMiddleware(jwtService, redisCache))
	{
		desktops.POST("/register", desktopHandler.RegisterDesktop)
		desktops.GET("", desktopHandler.ListDesktops)
		desktops.GET("/:id", desktopHandler.GetDesktop)
		desktops.PUT("/:id", desktopHandler.UpdateDesktop)
		desktops.DELETE("/:id", desktopHandler.DeleteDesktop)
		desktops.GET("/:id/status", desktopHandler.GetDesktopStatus)
	}

	// 会话相关（需要登录）
	sessions := v1.Group("/sessions")
	sessions.Use(middleware.AuthMiddleware(jwtService, redisCache))
	{
		sessions.POST("", sessionHandler.CreateSession)
		sessions.GET("", sessionHandler.ListSessions)
		sessions.GET("/:id", sessionHandler.GetSession)
		sessions.DELETE("/:id", sessionHandler.DeleteSession)
		sessions.GET("/:id/messages", sessionHandler.GetMessages)
	}

	// WebSocket 路由
	wsHandler.RegisterRoutes(router)
}
