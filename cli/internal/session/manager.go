package session

import (
	"encoding/base64"
	"fmt"
	"sync"

	"pocket-coder-cli/internal/terminal"
	"pocket-coder-cli/internal/websocket"
)

// Manager 管理多个终端会话
type Manager struct {
	mu          sync.RWMutex
	sessions    map[int64]*terminal.Terminal
	defaultTerm *terminal.Terminal
	wsClient    *websocket.Client
	workDir     string
}

// NewManager 创建会话管理器
func NewManager(wsClient *websocket.Client, defaultTerm *terminal.Terminal, workDir string) *Manager {
	return &Manager{
		sessions:    make(map[int64]*terminal.Terminal),
		defaultTerm: defaultTerm,
		wsClient:    wsClient,
		workDir:     workDir,
	}
}

// HandleSessionCreate 处理创建/分配会话
func (m *Manager) HandleSessionCreate(sessionID int64, workingDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查会话是否已存在
	if _, exists := m.sessions[sessionID]; exists {
		return
	}

	// 如果是第一个会话，绑定到默认终端（即用户看到的那个）
	// 注意：这里假设 Server 端会为每个连接的 Agent 发送至少一个 CreateSession 消息
	if len(m.sessions) == 0 {
		m.sessions[sessionID] = m.defaultTerm
		m.setupTerminalOutput(sessionID, m.defaultTerm)
		fmt.Printf("🔗 会话 #%d 已绑定到主终端\n", sessionID)
		return
	}

	// 否则，启动一个新的后台终端
	fmt.Printf("✨ 创建新终端会话 #%d...\n", sessionID)
	term := terminal.NewTerminal()
	
	// 不启用本地显示（只有主终端显示在 CLI 上）
	term.SetLocalDisplay(false)

	// 设置输出处理
	m.setupTerminalOutput(sessionID, term)

	// 启动终端
	dir := workingDir
	if dir == "" {
		dir = m.workDir
	}

	if err := term.Start(dir); err != nil {
		fmt.Printf("❌ 启动会话 #%d 失败: %v\n", sessionID, err)
		return
	}

	m.sessions[sessionID] = term
}

// setupTerminalOutput 设置终端的输出和退出处理
func (m *Manager) setupTerminalOutput(sessionID int64, term *terminal.Terminal) {
	// 输出转发
	term.OnOutput(func(data []byte) {
		encoded := base64.StdEncoding.EncodeToString(data)
		m.wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalOutput,
			Payload: map[string]interface{}{
				"session_id": sessionID,
				"data":       encoded,
			},
		})
	})

	// 退出处理
	term.OnExit(func(code int) {
		fmt.Printf("📤 会话 #%d 终端已退出 (code: %d)\n", sessionID, code)
		
		m.mu.Lock()
		delete(m.sessions, sessionID)
		isDefault := (term == m.defaultTerm)
		m.mu.Unlock()

		// 通知服务端会话结束（可选，目前服务端没有专门的 EndSession 消息接收逻辑，通常是通过 Agent 状态判断）
		// 但为了保持状态同步，可以发送一个 TypeTerminalExit
		m.wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalExit,
			Payload: map[string]interface{}{
				"session_id": sessionID,
				"code":       code,
			},
		})
		
		// 如果是默认终端退出，可能需要关闭整个程序？或者保持连接？
		// 目前保持连接
		if isDefault {
			fmt.Println("⚠️ 主终端已退出")
		}
	})
}

// Write 写入数据到指定会话
func (m *Manager) Write(sessionID int64, data []byte) error {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	// 如果没有指定 SessionID (0)，且只有一个会话，则使用该会话（兼容性）
	if !exists && sessionID == 0 && len(m.sessions) == 1 {
		for _, t := range m.sessions {
			term = t
			break
		}
		exists = true
	}
	m.mu.RUnlock()

	if !exists || term == nil {
		return fmt.Errorf("session %d not found", sessionID)
	}

	return term.Write(data)
}

// Resize 调整指定会话的大小
func (m *Manager) Resize(sessionID int64, rows, cols uint16) error {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	if !exists && sessionID == 0 && len(m.sessions) == 1 {
		for _, t := range m.sessions {
			term = t
			break
		}
		exists = true
	}
	m.mu.RUnlock()

	if !exists || term == nil {
		return fmt.Errorf("session %d not found", sessionID)
	}

	return term.Resize(rows, cols)
}

// GetHistory 获取指定会话的历史
func (m *Manager) GetHistory(sessionID int64) ([]byte, error) {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	if !exists && sessionID == 0 && len(m.sessions) == 1 {
		for _, t := range m.sessions {
			term = t
			break
		}
		exists = true
	}
	m.mu.RUnlock()

	if !exists || term == nil {
		return nil, fmt.Errorf("session %d not found", sessionID)
	}

	return term.GetHistory(), nil
}
