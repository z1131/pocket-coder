// Package cmd 实现 CLI 命令
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"pocket-coder-cli/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示当前状态",
	Long: `显示当前登录状态和配置信息。

包括：
- 服务器地址
- 登录状态
- 设备 ID（如果已登录）`,
	Run: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║           Pocket Coder 状态信息                 ║")
	fmt.Println("╠════════════════════════════════════════════════╣")

	// 服务器地址
	fmt.Printf("║  服务器: %s\n", config.GetServerURL())

	// 登录状态
	if config.GetAccessToken() != "" {
		fmt.Println("║  登录状态: ✓ 已登录")
		fmt.Printf("║  设备 ID: %s\n", config.GetDesktopID())
	} else {
		fmt.Println("║  登录状态: ✗ 未登录")
		fmt.Println("║")
		fmt.Println("║  请运行 'pocket-coder login' 完成登录")
	}

	fmt.Println("╚════════════════════════════════════════════════╝")
}
