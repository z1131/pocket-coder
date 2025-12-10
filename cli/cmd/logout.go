// Package cmd 实现 CLI 命令
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"pocket-coder-cli/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "登出并清除本地凭证",
	Long: `登出当前账号并清除本地保存的 token 和设备信息。

登出后需要重新运行 'pocket-coder login' 才能使用。`,
	Run: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) {
	// 检查是否已登录
	if config.GetAccessToken() == "" {
		fmt.Println("当前未登录")
		return
	}

	// 清除本地凭证
	if err := config.ClearToken(); err != nil {
		fmt.Fprintf(os.Stderr, "清除凭证失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ 已登出并清除本地凭证")
}
