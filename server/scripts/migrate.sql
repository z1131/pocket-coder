-- Pocket Coder 数据库迁移脚本
-- 用于手动创建数据库和表（可选，GORM AutoMigrate 也会自动创建）

-- 创建数据库
CREATE DATABASE IF NOT EXISTS pocket_coder
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE pocket_coder;

-- ============================================
-- 用户表
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    nickname VARCHAR(50) NOT NULL DEFAULT '' COMMENT '昵称',
    email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    password VARCHAR(100) NOT NULL COMMENT '密码（bcrypt加密）',
    user_code VARCHAR(20) NOT NULL UNIQUE COMMENT '用户唯一标识码',
    avatar VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_email (email),
    INDEX idx_user_code (user_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 电脑端设备表
-- ============================================
CREATE TABLE IF NOT EXISTS desktops (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '设备ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',
    name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '设备名称',
    hostname VARCHAR(255) NOT NULL DEFAULT '' COMMENT '主机名',
    os VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作系统',
    agent_version VARCHAR(20) NOT NULL DEFAULT '' COMMENT '客户端版本',
    last_seen_at DATETIME DEFAULT NULL COMMENT '最后在线时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_last_seen (last_seen_at),
    CONSTRAINT fk_desktop_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电脑端设备表';

-- ============================================
-- 会话表
-- ============================================
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '会话ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    desktop_id BIGINT NOT NULL COMMENT '设备ID',
    title VARCHAR(100) NOT NULL DEFAULT '' COMMENT '会话标题',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0=进行中，1=已完成，2=已归档',
    project_path VARCHAR(500) NOT NULL DEFAULT '' COMMENT '项目路径',
    message_count INT NOT NULL DEFAULT 0 COMMENT '消息数量',
    last_active_at DATETIME DEFAULT NULL COMMENT '最后活跃时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_user_desktop (user_id, desktop_id),
    INDEX idx_status (status),
    INDEX idx_last_active (last_active_at),
    CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_session_desktop FOREIGN KEY (desktop_id) REFERENCES desktops(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- ============================================
-- 消息表
-- ============================================
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '消息ID',
    session_id BIGINT NOT NULL COMMENT '会话ID',
    role VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '角色：user/assistant/system',
    content TEXT NOT NULL COMMENT '消息内容',
    content_type VARCHAR(20) NOT NULL DEFAULT 'text' COMMENT '内容类型：text/markdown/code/image/file',
    metadata JSON DEFAULT NULL COMMENT '附加元数据',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_session_id (session_id),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at),
    CONSTRAINT fk_message_session FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- ============================================
-- 初始化测试数据（可选）
-- ============================================
-- INSERT INTO users (nickname, email, password, user_code) VALUES 
-- ('测试用户', 'test@example.com', '$2a$10$...', 'USR_ABC123');
