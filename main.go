package main

import (
	"MyBlog/cmd"
	"MyBlog/internal/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "配置初始化失败: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "myblog",
	Short: "MyBlog - 简易静态博客系统",
	Long: `MyBlog 是一个简易的静态博客系统，类似于Hugo。
它专为命令行使用而设计，借助GitHub对Markdown文档的完美支持。

对我们来说，Markdown文档就是界面。`,
}

func init() {
	rootCmd.AddCommand(cmd.DraftCmd)
	rootCmd.AddCommand(cmd.PubCmd)
}
