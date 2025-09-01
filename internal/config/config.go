package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 配置结构体
type Config struct {
	Directories struct {
		Draft string `yaml:"draft"`
		Blogs string `yaml:"blogs"`
	} `yaml:"directories"`
}

var AppConfig *Config

// InitConfig 初始化配置
func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.myblog")

	// 设置默认值
	viper.SetDefault("directories.draft", "_draft")
	viper.SetDefault("directories.blogs", "blogs")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，创建默认配置文件
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := createDefaultConfig(); err != nil {
				return fmt.Errorf("创建默认配置文件失败: %v", err)
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 将配置解析到结构体
	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	configContent := `# MyBlog 配置文件
directories:
  draft: "_draft"
  blogs: "blogs"
`

	configFile := "config.yaml"
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return err
	}

	return viper.ReadInConfig()
}

// GetDraftDir 获取草稿目录
func GetDraftDir() string {
	if AppConfig != nil {
		return AppConfig.Directories.Draft
	}
	return "_draft"
}

// GetBlogsDir 获取博客目录
func GetBlogsDir() string {
	if AppConfig != nil {
		return AppConfig.Directories.Blogs
	}
	return "blogs"
}
