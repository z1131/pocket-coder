// Package session 管理多终端会话
package session

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"pocket-coder-cli/internal/terminal"
	"pocket-coder-cli/internal/websocket"
)

// Manager 管理多个终端会话
type Manager struct {
	mu            sync.RWMutex
	sessions      map[int64]*terminal.Terminal
	mainSessionID int64 // 主会话 ID (显示在本地终端)
	wsClient      *websocket.Client
	workDir       string
}

// NewManager 创建会话管理器
func NewManager(wsClient *websocket.Client, workDir string) *Manager {
	return &Manager{
		sessions: make(map[int64]*terminal.Terminal),
		wsClient: wsClient,
		workDir:  workDir,
	}
}

// HandleSessionCreate 处理创建/分配会话
func (m *Manager) HandleSessionCreate(sessionID int64, workingDir string, isDefault bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查会话是否已存在
	if _, exists := m.sessions[sessionID]; exists {
		return
	}

	term := terminal.NewTerminal()
	term.SetLocalDisplay(false) // 统一由 Manager 控制输出

	// 如果 Server 指定这是默认会话
	if isDefault {
		m.mainSessionID = sessionID
		fmt.Printf("\r\n🔗 默认终端会话 #%d 已连接\r\n", sessionID)
	}

	// 设置输出处理
	m.setupTerminalOutput(sessionID, term, isDefault)

	// 启动终端
	dir := workingDir
	if dir == "" {
		dir = m.workDir
	}

	if err := term.Start(dir); err != nil {
		if isDefault {
			fmt.Printf("❌ 启动默认会话 #%d 失败: %v\n", sessionID, err)
		}
		return
	}

	m.sessions[sessionID] = term
}

// HandleSessionClose 处理关闭会话
func (m *Manager) HandleSessionClose(sessionID int64) {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		// Stop 会 kill 掉 PTY 进程，触发 OnExit 回调
		term.Stop()
	}
}

// setupTerminalOutput 设置终端的输出和退出处理
func (m *Manager) setupTerminalOutput(sessionID int64, term *terminal.Terminal, isDefault bool) {
	// 输出转发
	term.OnOutput(func(data []byte) {
		// 1. 发送到 WebSocket (始终)
		encoded := base64.StdEncoding.EncodeToString(data)
		m.wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalOutput,
			Payload: map[string]interface{}{
				"session_id": sessionID,
				"data":       encoded,
			},
		})

		// 2. 如果是默认会话，写入本地 Stdout
		if isDefault {
			os.Stdout.Write(data)
		}
	})

	// 退出处理
	term.OnExit(func(code int) {
		m.mu.Lock()
		delete(m.sessions, sessionID)
		if m.mainSessionID == sessionID {
			m.mainSessionID = 0
		}
		m.mu.Unlock()

		// 通知服务端
		m.wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalExit,
			Payload: map[string]interface{}{
				"session_id": sessionID,
				"code":       code,
			},
		})
		
		if isDefault {
			fmt.Printf("\r\n📤 默认会话已退出 (code: %d)\r\n", code)
			// 默认会话退出通常意味着程序也该结束了，或者等待重连
			// 这里我们让 root.go 的逻辑来决定是否退出程序
			// 但为了安全，我们可以关闭所有后台会话
			m.Close()
			os.Exit(0) // 强制退出
		}
	})
}

// WriteToMain 写入数据到主会话（本地键盘输入）
func (m *Manager) WriteToMain(data []byte) error {
	m.mu.RLock()
	id := m.mainSessionID
	term := m.sessions[id]
	m.mu.RUnlock()
	
	if term == nil {
		return nil
	}
	return term.Write(data)
}

// Write 写入数据到指定会话（远程 WebSocket 输入）
func (m *Manager) Write(sessionID int64, data []byte) error {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	// 兼容旧逻辑：如果没传 ID，发给主会话
	if !exists && sessionID == 0 {
		term = m.sessions[m.mainSessionID]
		exists = (term != nil)
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
	defer m.mu.RUnlock()
	
	// 1. 如果 sessionID 为 0 (本地窗口变化)，只调整主会话
	if sessionID == 0 {
		if term, ok := m.sessions[m.mainSessionID]; ok {
			return term.Resize(rows, cols)
		}
		return nil
	}
	
	// 2. 远程调整指定会话
	if term, ok := m.sessions[sessionID]; ok {
		return term.Resize(rows, cols)
	}
	
	return fmt.Errorf("session %d not found", sessionID)
}

// Close 关闭所有会话
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, term := range m.sessions {
		term.Stop()
	}
	m.sessions = make(map[int64]*terminal.Terminal)
	m.mainSessionID = 0
}

// GetHistory 获取指定会话的历史
func (m *Manager) GetHistory(sessionID int64) ([]byte, error) {
	m.mu.RLock()
	term, exists := m.sessions[sessionID]
	if !exists && sessionID == 0 {
		term = m.sessions[m.mainSessionID]
		exists = (term != nil)
	}
	m.mu.RUnlock()

	if !exists || term == nil {
		return nil, fmt.Errorf("session %d not found", sessionID)
	}

	return term.GetHistory(), nil
}
