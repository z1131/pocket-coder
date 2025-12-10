// Package cmd 实现 CLI 命令
package cmd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"pocket-coder-cli/internal/api"
	"pocket-coder-cli/internal/config"
	"pocket-coder-cli/internal/terminal"
	"pocket-coder-cli/internal/websocket"
)

var rootCmd = &cobra.Command{
	Use:   "pocket-coder",
	Short: "Pocket Coder - 手机远程控制电脑端 AI 编程工具",
	Long: `Pocket Coder CLI 客户端

用于将手机端的指令转发给本地 AI 编程工具（如 Claude Code、Cursor 等）。

直接运行即可开始使用，程序会引导你完成登录和连接。`,
	Run: runInteractive,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局参数
	rootCmd.PersistentFlags().StringP("server", "s", "", "服务器地址 (默认: http://localhost:8080)")
}

func initConfig() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 如果指定了服务器地址，更新配置
	if server, _ := rootCmd.PersistentFlags().GetString("server"); server != "" {
		config.SetServerURL(server)
	}
}

// runInteractive 交互式主流程
func runInteractive(cmd *cobra.Command, args []string) {
	printBanner()

	// 检查是否已登录
	desktopToken := config.GetDesktopToken()
	if desktopToken != "" {
		fmt.Println("检测到已保存的登录信息")
		fmt.Printf("  设备 ID: %s\n", config.GetDesktopID())
		fmt.Println()

		if askYesNo("是否使用已保存的登录信息？") {
			startWebSocket()
			return
		}
		fmt.Println()
	}

	// 交互式登录
	doInteractiveLogin()

	// 登录成功后自动启动
	startWebSocket()
}

func printBanner() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║         🚀 Pocket Coder CLI 客户端              ║")
	fmt.Println("║                                                ║")
	fmt.Println("║   手机远程控制电脑端 AI 编程工具                  ║")
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()
}

func doInteractiveLogin() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("📱 开始登录")
	fmt.Println("─────────────────────────────────")
	fmt.Println()

	// 输入用户名
	fmt.Print("请输入用户名: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		fmt.Fprintln(os.Stderr, "✗ 用户名不能为空")
		os.Exit(1)
	}

	// 输入密码（隐藏输入）
	fmt.Print("请输入密码: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // 换行
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 读取密码失败: %v\n", err)
		os.Exit(1)
	}
	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		fmt.Fprintln(os.Stderr, "✗ 密码不能为空")
		os.Exit(1)
	}

	fmt.Println()

	// 登录
	client := api.NewClient(config.GetServerURL())

	fmt.Println("🔐 正在登录...")
	loginResp, err := client.Login(username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 登录失败: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveAuth(loginResp.AccessToken, loginResp.RefreshToken); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 保存登录信息失败: %v\n", err)
		os.Exit(1)
	}

	// 注册/绑定桌面
	hostname := getHostname()
	osInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	workingDir, _ := os.Getwd()
	agentType := "claude-code"

	fmt.Println("💻 正在绑定当前电脑...")
	regReq := &api.RegisterDesktopRequest{
		Name:       hostname,
		AgentType:  &agentType,
		WorkingDir: &workingDir,
		OSInfo:     &osInfo,
	}

	regResp, err := client.RegisterDesktop(loginResp.AccessToken, regReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 绑定电脑失败: %v\n", err)
		os.Exit(1)
	}

	desktopIDStr := fmt.Sprintf("%d", regResp.DesktopID)
	if err := config.SaveDesktop(regResp.DesktopToken, desktopIDStr, regResp.Name); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 保存桌面信息失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ 登录并绑定成功！")
	fmt.Println("─────────────────────────────────")
	fmt.Printf("  👤 账号: %s\n", username)
	fmt.Printf("  🖥️  设备: %s (ID: %d)\n", regResp.Name, regResp.DesktopID)
	fmt.Printf("  📁 工作目录: %s\n", filepath.Clean(workingDir))
	fmt.Println()
}

func startWebSocket() {
	desktopToken := config.GetDesktopToken()
	desktopID := config.GetDesktopID()

	if desktopToken == "" || desktopID == "" {
		fmt.Fprintln(os.Stderr, "✗ 登录信息不完整，请重新登录")
		os.Exit(1)
	}

	workDir, _ := os.Getwd()

	fmt.Println("🌐 正在连接服务器...")
	fmt.Println("─────────────────────────────────")
	fmt.Printf("  📡 服务器: %s\n", config.GetServerURL())
	fmt.Printf("  🔑 设备 ID: %s\n", desktopID)
	fmt.Printf("  📁 工作目录: %s\n", workDir)
	fmt.Println()

	// 创建 PTY 终端
	ptyTerm := terminal.NewTerminal()
	
	// 启用本地显示
	ptyTerm.SetLocalDisplay(true)

	// 创建 WebSocket 客户端
	wsClient := websocket.NewClient(config.GetServerURL(), desktopToken, desktopID)

	// 设置消息处理
	setupTerminalHandlers(wsClient, ptyTerm, workDir)

	// 连接服务器
	if err := wsClient.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 连接服务器失败: %v\n", err)
		os.Exit(1)
	}

	// 启动终端
	if err := ptyTerm.Start(workDir); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 启动终端失败: %v\n", err)
		wsClient.Disconnect()
		os.Exit(1)
	}

	fmt.Println("✅ 已连接到服务器！")
	fmt.Println("✅ 终端已启动！")
	fmt.Println()
	fmt.Println("📱 手机端和电脑端可以同时操作此终端")
	fmt.Println("   (按 Ctrl+\\ 退出)")
	fmt.Println()
	fmt.Println("─────────────────────────────────")
	fmt.Println()

	// 将当前终端设为 raw mode，以便捕获所有按键
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 设置终端 raw mode 失败: %v\n", err)
		ptyTerm.Stop()
		wsClient.Disconnect()
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 设置终端大小为当前窗口大小
	if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		ptyTerm.Resize(uint16(height), uint16(width))
	}

	// 用于控制退出的 channel
	done := make(chan struct{})

	// 启动本地键盘输入读取
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
			}

			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				// 检查是否按下 Ctrl+\ (0x1c) 退出
				for i := 0; i < n; i++ {
					if buf[i] == 0x1c {
						close(done)
						return
					}
				}

				// 写入 PTY
				ptyTerm.Write(buf[:n])
			}
		}
	}()

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
	case <-done:
	}

	// 恢复终端状态（defer 会处理）
	fmt.Println()
	fmt.Println()
	fmt.Println("正在断开连接...")

	// 停止终端
	ptyTerm.Stop()

	// 断开 WebSocket
	wsClient.Disconnect()

	fmt.Println("✅ 已断开连接，再见！")
}

// setupTerminalHandlers 设置终端消息处理器
func setupTerminalHandlers(wsClient *websocket.Client, term *terminal.Terminal, workDir string) {
	// 终端输出 → 发送到手机端
	term.OnOutput(func(data []byte) {
		// 使用 base64 编码二进制数据
		encoded := base64.StdEncoding.EncodeToString(data)
		wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalOutput,
			Payload: map[string]interface{}{
				"data": encoded,
			},
		})
	})

	// 终端退出
	term.OnExit(func(code int) {
		fmt.Printf("📤 终端已退出 (code: %d)\n", code)
		wsClient.SendMessage(&websocket.Message{
			Type: websocket.TypeTerminalExit,
			Payload: map[string]interface{}{
				"code": code,
			},
		})
	})

	// 处理来自服务器的消息
	wsClient.OnMessage(func(msg *websocket.Message) {
		switch msg.Type {
		case websocket.TypeTerminalInput, "user:message":
			// 手机端输入
			handleTerminalInput(term, msg)

		case websocket.TypeTerminalResize:
			// 调整终端大小
			handleTerminalResize(term, msg)

		case "ping":
			// 心跳响应
			wsClient.SendMessage(&websocket.Message{
				Type: "pong",
			})

		case "pong":
			// 忽略心跳响应
			return

		default:
			fmt.Printf("⚠️  未知消息类型: %s\n", msg.Type)
		}
	})
}

// handleTerminalInput 处理终端输入
func handleTerminalInput(term *terminal.Terminal, msg *websocket.Message) {
	var data string

	// 从 payload 获取数据
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		if d, ok := payload["data"].(string); ok {
			data = d
		}
	}

	// 兼容旧格式：从 content 获取
	if data == "" && msg.Content != "" {
		data = msg.Content
	}

	if data == "" {
		return
	}

	// 尝试 base64 解码，如果失败则当作纯文本
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		// 不是 base64，当作纯文本
		decoded = []byte(data)
	}

	// 调试日志（输出到 stderr 避免干扰终端）
	// fmt.Fprintf(os.Stderr, "[DEBUG] 收到手机输入: %q (解码后: %q)\n", data, string(decoded))

	// 写入终端
	if err := term.Write(decoded); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 写入终端失败: %v\n", err)
	}
}

// handleTerminalResize 处理终端大小调整
func handleTerminalResize(term *terminal.Terminal, msg *websocket.Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	rows, _ := payload["rows"].(float64)
	cols, _ := payload["cols"].(float64)

	if rows > 0 && cols > 0 {
		if err := term.Resize(uint16(rows), uint16(cols)); err != nil {
			fmt.Printf("❌ 调整终端大小失败: %v\n", err)
		} else {
			fmt.Printf("📐 终端大小调整为 %dx%d\n", int(cols), int(rows))
		}
	}
}

func askYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n]: ", prompt)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "unknown"
	}
	return hostname
}
