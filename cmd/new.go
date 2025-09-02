package cmd

import (
	"MyBlog/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	newTags       []string
	newTagsString string
	newVerbose    bool
)

var NewCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "创建一篇新的正式文章",
	Long: `创建一篇新的正式文章到blogs目录中。

文章将按照标签创建目录结构，目录路径可在配置文件中自定义。
文章会自动添加发布时间。如果未提供文章标题，将会启动交互式模式来收集必要信息。`,
	Example: `  myblog new "我的第一篇博客" --tags "Go/基础"
  myblog new "设计模式实践" --tags "Go/设计模式/教程"
  myblog new  # 交互式模式`,
	Args: cobra.MaximumNArgs(1),
	Run:  runNewCommand,
}

func init() {
	// 添加命令行标志
	NewCmd.Flags().StringVarP(&newTagsString, "tags", "t", "", "文章标签路径 (使用斜杠分隔创建目录结构，如: Go/基础/教程)")
	NewCmd.Flags().BoolVarP(&newVerbose, "verbose", "v", false, "详细输出")

	// 设置日志级别
	if newVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func runNewCommand(cmd *cobra.Command, args []string) {
	// 设置颜色输出
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	var title string

	// 获取文章标题
	if len(args) > 0 {
		title = args[0]
		// 处理命令行标签参数
		if newTagsString != "" {
			tags := strings.Split(newTagsString, "/")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					newTags = append(newTags, tag)
				}
			}
		}
	} else {
		// 交互式获取信息
		articleInfo, err := getNewArticleInfoInteractively()
		if err != nil {
			fmt.Printf("%s %v\n", red("错误:"), err)
			return
		}
		title = articleInfo.Title
		newTags = articleInfo.Tags
	}

	if title == "" {
		fmt.Printf("%s 文章标题不能为空\n", red("错误:"))
		return
	}

	fmt.Printf("%s 正在创建正式文章: %s\n", blue("信息:"), yellow(title))

	// 创建正式文章
	filePath, err := createNewArticle(title, newTags)
	if err != nil {
		fmt.Printf("%s 创建文章失败: %v\n", red("错误:"), err)
		logrus.WithError(err).Error("创建文章失败")
		return
	}

	fmt.Printf("%s 成功创建正式文章!\n", green("✓"))
	fmt.Printf("  文件路径: %s\n", green(filePath))
	fmt.Printf("  标题: %s\n", title)
	if len(newTags) > 0 {
		fmt.Printf("  标签: %s\n", strings.Join(newTags, "/"))
		fmt.Printf("  目录结构: %s\n", blue(strings.Join(newTags, "/")))
	}
	fmt.Printf("  发布时间: %s\n", time.Now().Format("2006年01月02日 15:04"))

	logrus.WithFields(logrus.Fields{
		"title": title,
		"path":  filePath,
		"tags":  newTags,
		"type":  "published",
	}).Info("正式文章创建成功")
}

type NewArticleInfo struct {
	Title string
	Tags  []string
}

func getNewArticleInfoInteractively() (*NewArticleInfo, error) {
	info := &NewArticleInfo{}

	// 获取已有标签路径用于选择
	existingTagPaths := getExistingTagsForNew()

	// 1. 获取文章标题
	titleQuestion := &survey.Input{
		Message: "请输入文章标题:",
		Help:    "这将是你文章的主标题",
	}

	err := survey.AskOne(titleQuestion, &info.Title, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, err
	}

	// 2. 选择标签路径方式
	var tagChoice string
	if len(existingTagPaths) > 0 {
		// 添加"输入新标签"选项
		options := append([]string{"输入新标签路径"}, existingTagPaths...)

		tagSelectQuestion := &survey.Select{
			Message: "请选择标签路径:",
			Options: options,
			Help:    "选择现有的标签路径，或选择'输入新标签路径'来创建新的目录结构",
		}

		err = survey.AskOne(tagSelectQuestion, &tagChoice)
		if err != nil {
			return nil, err
		}
	} else {
		tagChoice = "输入新标签路径"
	}

	// 3. 处理标签
	var finalTags []string
	if tagChoice == "输入新标签路径" {
		// 输入自定义标签路径
		var customTagsInput string
		customTagQuestion := &survey.Input{
			Message: "请输入标签路径 (使用斜杠分隔创建多级目录):",
			Help:    fmt.Sprintf("例如: Go/设计模式/单例 → %s/Go/设计模式/单例/", config.GetBlogsDir()),
		}

		err = survey.AskOne(customTagQuestion, &customTagsInput)
		if err != nil {
			return nil, err
		}

		if customTagsInput != "" {
			// 使用斜杠分隔标签路径
			tags := strings.Split(customTagsInput, "/")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					finalTags = append(finalTags, tag)
				}
			}
		}
	} else {
		// 使用选择的现有标签路径
		finalTags = strings.Split(tagChoice, "/")
	}

	info.Tags = finalTags

	// 显示最终的目录结构预览
	if len(finalTags) > 0 {
		fmt.Printf("\n📁 目录结构预览: %s → %s/%s/\n",
			strings.Join(finalTags, "/"),
			config.GetBlogsDir(),
			strings.Join(finalTags, "/"))
	} else {
		fmt.Printf("\n📁 目录结构预览: → %s/\n", config.GetBlogsDir())
	}

	return info, nil
}

// 获取已存在的完整标签路径（主要从blogs目录，也扫描_draft目录作为参考）
func getExistingTagsForNew() []string {
	tagPaths := make(map[string]bool)

	// 主要扫描blogs目录
	scanTagPathsRecursively(config.GetBlogsDir(), "", tagPaths)

	// 也扫描_draft目录作为参考
	scanTagPathsRecursively(config.GetDraftDir(), "", tagPaths)

	// 转换为切片并排序
	result := make([]string, 0, len(tagPaths))
	for tagPath := range tagPaths {
		if tagPath != "" {
			result = append(result, tagPath)
		}
	}

	return result
}

func createNewArticle(title string, tags []string) (string, error) {
	// 构建目录路径
	var dirPath string
	if len(tags) > 0 {
		// 使用标签作为目录结构：blogs/tag1/tag2/...
		tagPath := strings.Join(tags, string(filepath.Separator))
		dirPath = filepath.Join(config.GetBlogsDir(), tagPath)
	} else {
		// 如果没有标签，直接放在blogs目录下
		dirPath = config.GetBlogsDir()
	}

	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名
	fileName := sanitizeFileNameForNew(title) + ".md"
	filePath := filepath.Join(dirPath, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("文件已存在: %s", filePath)
	}

	// 创建文件内容
	content := generateNewMarkdownContent(title, tags)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	return filePath, nil
}

func sanitizeFileNameForNew(title string) string {
	// 将标题转换为适合作为文件名的格式
	fileName := strings.ToLower(title)

	// 替换空格和特殊字符为下划线
	specialChars := []string{" ", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "。", "，", "！", "？", "；", "："}
	for _, char := range specialChars {
		fileName = strings.ReplaceAll(fileName, char, "_")
	}

	// 移除连续的下划线
	for strings.Contains(fileName, "__") {
		fileName = strings.ReplaceAll(fileName, "__", "_")
	}

	// 移除开头和结尾的下划线
	fileName = strings.Trim(fileName, "_")

	// 如果文件名为空，使用默认名称
	if fileName == "" {
		fileName = "untitled"
	}

	return fileName
}

func generateNewMarkdownContent(title string, tags []string) string {
	now := time.Now()

	// 构建标签数组字符串
	var tagStr string
	if len(tags) > 0 {
		tagList := make([]string, len(tags))
		for i, tag := range tags {
			tagList[i] = fmt.Sprintf(`"%s"`, tag)
		}
		tagStr = strings.Join(tagList, ", ")
	}

	content := fmt.Sprintf(`---
title: "%s"
date: %s
published: %s
tags: [%s]
---

# %s

在这里开始写你的文章内容...

## 介绍

简要介绍文章的主要内容。

## 主要内容

更多内容...

## 总结

总结你的观点和主要内容。

---

> 发布时间: %s
`, title, now.Format("2006-01-02T15:04:05Z07:00"), now.Format("2006-01-02T15:04:05Z07:00"), tagStr, title, now.Format("2006年01月02日 15:04"))

	return content
}
