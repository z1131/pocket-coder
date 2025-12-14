// Package cmd 实现 CLI 命令
package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"pocket-coder-cli/internal/api"
	"pocket-coder-cli/internal/config"
	"pocket-coder-cli/internal/session"
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
		fmt.Println("检测到有效的登录凭证：")
		username := config.Get().Device.Username
		if username == "" {
			username = "未知用户"
		}
		fmt.Printf("  👤 用户: %s\n", username)
		fmt.Printf("  💻 设备: %s (ID: %s)\n", config.Get().Device.Name, config.GetDesktopID())
		fmt.Printf("  🌐 服务器: %s\n", config.GetServerURL())
		fmt.Println()

		if askYesNo("是否直接连接？") {
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

	if err := config.SaveAuth(loginResp.AccessToken, loginResp.RefreshToken, loginResp.User.Username); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 保存登录信息失败: %v\n", err)
		os.Exit(1)
	}

	// 注册/绑定桌面
	hostname := getHostname()
	osInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// 获取或生成设备 UUID
	deviceUUID, err := config.GetDeviceUUID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 获取设备标识失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("💻 正在绑定当前电脑...")
	regReq := &api.RegisterDesktopRequest{
		Name:       hostname,
		DeviceUUID: deviceUUID,
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

	// 生成 Process ID (随机 UUID 风格)
	b := make([]byte, 16)
	rand.Read(b)
	processID := hex.EncodeToString(b)

	// 创建 WebSocket 客户端
	wsClient := websocket.NewClient(config.GetServerURL(), desktopToken, desktopID, processID)

	// 创建会话管理器
	sessMgr := session.NewManager(wsClient, workDir)

	// 设置消息处理
	setupHandlers(wsClient, sessMgr)

	// 监听断开重连
	reconnectChan := make(chan struct{}, 1)
	wsClient.OnClose(func() {
		select {
		case reconnectChan <- struct{}{}:
		default:
		}
	})

	// 连接服务器
	if err := wsClient.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ 连接服务器失败: %v\n", err)
		os.Exit(1)
	}

	// 将当前终端设为 raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ 设置终端 raw mode 失败: %v\n", err)
		wsClient.Disconnect()
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

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
				var dataToSend []byte
				
				for i := 0; i < n; i++ {
					b := buf[i]
					if b == 0x1c { // Ctrl+\ (ASCII 28) -> 退出
						close(done)
						return
					} else {
						dataToSend = append(dataToSend, b)
					}
				}

				if len(dataToSend) > 0 {
					sessMgr.WriteToMain(dataToSend)
				}
			}
		}
	}()

	// 监听窗口大小变化
	if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		sessMgr.Resize(0, uint16(height), uint16(width))
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 主事件循环
	loop:
	for {
		select {
		case <-sigChan:
			break loop
		case <-done:
			break loop
		case <-reconnectChan:
			// 连接断开，尝试重连
			// 暂时恢复终端状态以便打印日志
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Println("\r\n⚠️  连接断开，3秒后尝试重连...")
			
			// 重试循环
			for {
				time.Sleep(3 * time.Second)
				
				// 检查是否已退出
				select {
				case <-sigChan:
					break loop
				case <-done:
					break loop
				default:
				}

				fmt.Print("🔄 正在重连... ")
				if err := wsClient.Connect(); err != nil {
					fmt.Printf("失败: %v\n", err)
				} else {
					fmt.Println("成功！")
					// 恢复 Raw Mode
					term.MakeRaw(int(os.Stdin.Fd()))
					
					// 发送 Resize 以同步状态
					if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
						sessMgr.Resize(0, uint16(height), uint16(width))
					}
					break // 重连成功，回到主循环
				}
			}
		}
	}

	// 恢复终端状态
	term.Restore(int(os.Stdin.Fd()), oldState)
	
	// 关闭所有会话
	sessMgr.Close()
	
	fmt.Println()
	fmt.Println("正在断开连接...")

	// 断开 WebSocket
	wsClient.Disconnect()

	fmt.Println("✅ 已断开连接，再见！")
}

// setupHandlers 设置 WebSocket 消息处理器
func setupHandlers(wsClient *websocket.Client, sessMgr *session.Manager) {
	wsClient.OnMessage(func(msg *websocket.Message) {
		switch msg.Type {
		case websocket.TypeSessionCreate:
			// 创建/分配会话
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				var sessionID int64
				if sid, ok := payload["session_id"].(float64); ok {
					sessionID = int64(sid)
				}
				workingDir, _ := payload["working_dir"].(string)
				isDefault, _ := payload["is_default"].(bool) // 字段名变更

				if sessionID > 0 {
					sessMgr.HandleSessionCreate(sessionID, workingDir, isDefault)
				}
			}

		case websocket.TypeSessionClose:
			// 关闭会话
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				var sessionID int64
				if sid, ok := payload["session_id"].(float64); ok {
					sessionID = int64(sid)
				}

				if sessionID > 0 {
					sessMgr.HandleSessionClose(sessionID)
				}
			}

		case websocket.TypeTerminalInput, "user:message":
			// 手机端输入
			var sessionID int64
			var data string

			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				if sid, ok := payload["session_id"].(float64); ok {
					sessionID = int64(sid)
				}
				if d, ok := payload["data"].(string); ok {
					data = d
				}
			}

			// 兼容旧格式
			if data == "" && msg.Content != "" {
				data = msg.Content
			}

			if data != "" {
				// Base64 解码
				decoded, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					decoded = []byte(data)
				}
				sessMgr.Write(sessionID, decoded)
			}

		case websocket.TypeTerminalResize:
			// 调整终端大小
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				var sessionID int64
				if sid, ok := payload["session_id"].(float64); ok {
					sessionID = int64(sid)
				}
			
rows, _ := payload["rows"].(float64)
cols, _ := payload["cols"].(float64)

				if rows > 0 && cols > 0 {
					sessMgr.Resize(sessionID, uint16(rows), uint16(cols))
				}
			}

		case websocket.TypeTerminalHistory:
			// 请求历史记录
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				var sessionID int64
				if sid, ok := payload["session_id"].(float64); ok {
					sessionID = int64(sid)
				}
				
				history, err := sessMgr.GetHistory(sessionID)
				if err == nil && len(history) > 0 {
					encoded := base64.StdEncoding.EncodeToString(history)
					wsClient.SendMessage(&websocket.Message{
						Type: websocket.TypeTerminalHistory,
						Payload: map[string]interface{}{
							"session_id": sessionID,
							"data":       encoded,
						},
					})
				}
			}

		case "ping":
			wsClient.SendMessage(&websocket.Message{Type: "pong"})
		}
	})
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