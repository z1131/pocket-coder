// Package terminal 提供 PTY 终端桥接功能
package terminal

import (
"fmt"
"io"
"os"
"os/exec"
"sync"

"github.com/creack/pty"
)

// Terminal PTY 终端
type Terminal struct {
	cmd          *exec.Cmd
	pty          *os.File
	mu           sync.Mutex
	isRunning    bool
	onOutput     func(data []byte)
	onExit       func(code int)
	done         chan struct{}
	rows         uint16
	cols         uint16
	localDisplay bool // 是否在本地终端显示输出
}

// NewTerminal 创建新终端
func NewTerminal() *Terminal {
	return &Terminal{
		rows:         24,
		cols:         80,
		done:         make(chan struct{}),
		localDisplay: false,
	}
}

// SetLocalDisplay 设置是否在本地终端显示输出
func (t *Terminal) SetLocalDisplay(enable bool) {
	t.localDisplay = enable
}

// OnOutput 设置输出回调
func (t *Terminal) OnOutput(handler func(data []byte)) {
	t.onOutput = handler
}

// OnExit 设置退出回调
func (t *Terminal) OnExit(handler func(code int)) {
	t.onExit = handler
}

// Start 启动终端
func (t *Terminal) Start(workingDir string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		return fmt.Errorf("终端已在运行")
	}

	// 获取用户的默认 shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// 创建命令
	t.cmd = exec.Command(shell)
	t.cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	if workingDir != "" {
		t.cmd.Dir = workingDir
	}

	// 启动 PTY
	ptmx, err := pty.StartWithSize(t.cmd, &pty.Winsize{
		Rows: t.rows,
		Cols: t.cols,
	})
	if err != nil {
		return fmt.Errorf("启动 PTY 失败: %w", err)
	}

	t.pty = ptmx
	t.isRunning = true
	t.done = make(chan struct{})

	// 启动输出读取
	go t.readOutput()

	// 监控进程退出
	go t.waitExit()

	return nil
}

// Stop 停止终端
func (t *Terminal) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		return nil
	}

	t.isRunning = false

	// 关闭 done channel
	select {
	case <-t.done:
	default:
		close(t.done)
	}

	// 关闭 PTY
	if t.pty != nil {
		t.pty.Close()
	}

	// 终止进程
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}

	return nil
}

// Write 写入数据到终端
func (t *Terminal) Write(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning || t.pty == nil {
		return fmt.Errorf("终端未运行")
	}

	_, err := t.pty.Write(data)
	return err
}

// Resize 调整终端大小
func (t *Terminal) Resize(rows, cols uint16) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = rows
	t.cols = cols

	if t.pty == nil {
		return nil
	}

	return pty.Setsize(t.pty, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

// IsRunning 检查是否运行中
func (t *Terminal) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isRunning
}

// readOutput 读取 PTY 输出
func (t *Terminal) readOutput() {
	buf := make([]byte, 4096)

	for {
		select {
		case <-t.done:
			return
		default:
		}

		n, err := t.pty.Read(buf)
		if err != nil {
			if err != io.EOF {
				// 非正常结束时记录
				// log.Printf("PTY 读取错误: %v", err)
			}
			return
		}

		if n > 0 {
			// 复制数据以避免被覆盖
			data := make([]byte, n)
			copy(data, buf[:n])

			// 如果启用本地显示，同时输出到标准输出
			if t.localDisplay {
				os.Stdout.Write(data)
			}

			// 回调
			if t.onOutput != nil {
				t.onOutput(data)
			}
		}
	}
}

// waitExit 等待进程退出
func (t *Terminal) waitExit() {
	if t.cmd == nil {
		return
	}

	err := t.cmd.Wait()

	t.mu.Lock()
	t.isRunning = false
	t.mu.Unlock()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	if t.onExit != nil {
		t.onExit(exitCode)
	}
}
