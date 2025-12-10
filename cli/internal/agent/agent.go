// Package agent 处理与 AI 编程工具的交互
package agent

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Agent AI 工具代理接口
type Agent interface {
	// Start 启动代理
	Start() error
	// Stop 停止代理
	Stop() error
	// Send 发送消息给 AI
	Send(message string) error
	// OnResponse 设置响应回调
	OnResponse(handler func(chunk string, isEnd bool))
	// IsRunning 检查是否运行中
	IsRunning() bool
}

// ClaudeCodeAgent Claude Code CLI 代理
type ClaudeCodeAgent struct {
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	onResponse    func(chunk string, isEnd bool)
	isRunning     bool
	mu            sync.Mutex
	workingDir    string
}

// NewClaudeCodeAgent 创建 Claude Code 代理
func NewClaudeCodeAgent(workingDir string) *ClaudeCodeAgent {
	return &ClaudeCodeAgent{
		workingDir: workingDir,
	}
}

// Start 启动 Claude Code
func (a *ClaudeCodeAgent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("代理已在运行")
	}

	// 启动 claude 命令（交互模式）
	a.cmd = exec.Command("claude", "--interactive")
	if a.workingDir != "" {
		a.cmd.Dir = a.workingDir
	}

	var err error

	// 获取标准输入
	a.stdin, err = a.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取 stdin 失败: %w", err)
	}

	// 获取标准输出
	a.stdout, err = a.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取 stdout 失败: %w", err)
	}

	// 获取标准错误
	a.stderr, err = a.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取 stderr 失败: %w", err)
	}

	// 启动进程
	if err := a.cmd.Start(); err != nil {
		return fmt.Errorf("启动 Claude Code 失败: %w", err)
	}

	a.isRunning = true

	// 启动输出读取协程
	go a.readOutput()

	return nil
}

// Stop 停止代理
func (a *ClaudeCodeAgent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return nil
	}

	a.isRunning = false

	if a.stdin != nil {
		a.stdin.Close()
	}

	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
	}

	return nil
}

// Send 发送消息
func (a *ClaudeCodeAgent) Send(message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("代理未运行")
	}

	// 发送消息到 Claude Code
	_, err := fmt.Fprintf(a.stdin, "%s\n", message)
	return err
}

// OnResponse 设置响应回调
func (a *ClaudeCodeAgent) OnResponse(handler func(chunk string, isEnd bool)) {
	a.onResponse = handler
}

// IsRunning 检查是否运行中
func (a *ClaudeCodeAgent) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isRunning
}

// readOutput 读取输出
func (a *ClaudeCodeAgent) readOutput() {
	scanner := bufio.NewScanner(a.stdout)
	var buffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// 检测响应结束标记（根据 Claude Code 的输出格式调整）
		if strings.HasPrefix(line, "claude>") || strings.HasPrefix(line, ">") {
			// 响应结束
			if buffer.Len() > 0 && a.onResponse != nil {
				a.onResponse(buffer.String(), true)
				buffer.Reset()
			}
		} else {
			// 流式输出
			buffer.WriteString(line)
			buffer.WriteString("\n")
			if a.onResponse != nil {
				a.onResponse(line+"\n", false)
			}
		}
	}
}

// MockAgent 模拟代理（用于测试）
type MockAgent struct {
	onResponse func(chunk string, isEnd bool)
	isRunning  bool
	mu         sync.Mutex
}

// NewMockAgent 创建模拟代理
func NewMockAgent() *MockAgent {
	return &MockAgent{}
}

// Start 启动
func (a *MockAgent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isRunning = true
	return nil
}

// Stop 停止
func (a *MockAgent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isRunning = false
	return nil
}

// Send 发送消息（模拟响应）
func (a *MockAgent) Send(message string) error {
	a.mu.Lock()
	if !a.isRunning {
		a.mu.Unlock()
		return fmt.Errorf("代理未运行")
	}
	a.mu.Unlock()

	// 模拟 AI 响应
	go func() {
		if a.onResponse != nil {
			response := fmt.Sprintf("收到消息: %s\n这是一个模拟响应。", message)
			// 模拟流式输出
			for _, char := range response {
				a.onResponse(string(char), false)
			}
			a.onResponse("", true)
		}
	}()

	return nil
}

// OnResponse 设置回调
func (a *MockAgent) OnResponse(handler func(chunk string, isEnd bool)) {
	a.onResponse = handler
}

// IsRunning 检查状态
func (a *MockAgent) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isRunning
}
