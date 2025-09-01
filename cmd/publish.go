package cmd

import (
	"MyBlog/internal/config"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var PubCmd = &cobra.Command{
	Use:   "pub [path]",
	Short: "发布草稿文章到博客目录",
	Long: `将草稿文章从草稿目录迁移到博客目录，并更新文章的发布时间。

支持以下发布方式：
1. 按文章路径发布：提供相对于草稿目录的路径
2. 交互式选择发布：不提供参数时进入交互模式

发布后会保持原有的目录结构，并更新文章末尾的更新时间。`,
	Example: `  myblog pub "Go/设计模式/实践/go设计模式实践.md"  # 按路径发布
  myblog pub                                        # 交互式选择草稿`,
	Args: cobra.MaximumNArgs(1),
	Run:  runPublishCommand,
}

func init() {
	PubCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
}

func runPublishCommand(cmd *cobra.Command, args []string) {
	// 设置颜色输出
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	var selectedDraft string

	// 获取草稿文件
	if len(args) > 0 {
		// 按路径查找草稿
		draftFile, err := findDraftByPath(args[0])
		if err != nil {
			fmt.Printf("%s %v\n", red("错误:"), err)
			return
		}
		selectedDraft = draftFile
	} else {
		// 交互式选择草稿
		draftFile, err := selectDraftInteractively()
		if err != nil {
			fmt.Printf("%s %v\n", red("错误:"), err)
			return
		}
		selectedDraft = draftFile
	}

	if selectedDraft == "" {
		fmt.Printf("%s 没有找到要发布的草稿\n", red("错误:"))
		return
	}

	fmt.Printf("%s 正在发布草稿: %s\n", blue("信息:"), yellow(filepath.Base(selectedDraft)))

	// 发布草稿
	publishedPath, err := publishDraft(selectedDraft)
	if err != nil {
		fmt.Printf("%s 发布失败: %v\n", red("错误:"), err)
		logrus.WithError(err).Error("发布草稿失败")
		return
	}

	fmt.Printf("%s 成功发布草稿!\n", green("✓"))
	fmt.Printf("  原路径: %s\n", selectedDraft)
	fmt.Printf("  新路径: %s\n", green(publishedPath))
	fmt.Printf("  发布时间: %s\n", time.Now().Format("2006年01月02日 15:04"))

	logrus.WithFields(logrus.Fields{
		"original_path":  selectedDraft,
		"published_path": publishedPath,
		"publish_time":   time.Now(),
	}).Info("草稿发布成功")
}

// 按路径查找草稿文件

// 按路径查找草稿文件
func findDraftByPath(inputPath string) (string, error) {
	// 规范化路径分隔符
	inputPath = filepath.FromSlash(inputPath)

	var fullPath string

	// 判断输入路径是否包含草稿目录前缀
	draftDir := config.GetDraftDir()
	if strings.HasPrefix(inputPath, draftDir+string(filepath.Separator)) ||
		strings.HasPrefix(inputPath, "./"+draftDir+string(filepath.Separator)) ||
		strings.HasPrefix(inputPath, ".\\"+draftDir+string(filepath.Separator)) {
		// 如果包含草稿目录前缀，直接使用该路径
		fullPath = inputPath
		// 去掉可能的 "./" 或 ".\\" 前缀
		if strings.HasPrefix(fullPath, "./") || strings.HasPrefix(fullPath, ".\\") {
			fullPath = fullPath[2:]
		}
	} else {
		// 如果不包含草稿目录前缀，拼接草稿目录
		fullPath = filepath.Join(draftDir, inputPath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("草稿文件不存在: %s", inputPath)
		}
		return "", fmt.Errorf("访问文件失败: %v", err)
	}

	// 验证是否为 Markdown 文件
	if !strings.HasSuffix(strings.ToLower(fullPath), ".md") {
		return "", fmt.Errorf("指定的文件不是 Markdown 文件: %s", inputPath)
	}

	// 验证文件确实在草稿目录中
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("获取绝对路径失败: %v", err)
	}

	absDraftDir, err := filepath.Abs(draftDir)
	if err != nil {
		return "", fmt.Errorf("获取草稿目录绝对路径失败: %v", err)
	}

	if !strings.HasPrefix(absFullPath, absDraftDir+string(filepath.Separator)) {
		return "", fmt.Errorf("指定文件不在草稿目录中: %s", inputPath)
	}

	return absFullPath, nil
}

// 交互式选择草稿
func selectDraftInteractively() (string, error) {
	drafts := getAllDrafts()

	if len(drafts) == 0 {
		return "", fmt.Errorf("草稿目录中没有找到任何文章")
	}

	options := make([]string, len(drafts))
	for i, draft := range drafts {
		relPath, _ := filepath.Rel(config.GetDraftDir(), draft)
		title := extractTitleFromFile(draft)
		if title != "" {
			options[i] = fmt.Sprintf("%s (%s)", title, relPath)
		} else {
			options[i] = relPath
		}
	}

	var selectedIndex int
	prompt := &survey.Select{
		Message: "请选择要发布的草稿:",
		Options: options,
		Help:    "选择一篇草稿文章发布到博客目录",
	}

	if err := survey.AskOne(prompt, &selectedIndex); err != nil {
		return "", err
	}

	return drafts[selectedIndex], nil
}

// 从文件中提取标题
func extractTitleFromFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontMatter := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				break
			}
		}

		if inFrontMatter && strings.HasPrefix(line, "title:") {
			titleValue := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			return strings.Trim(titleValue, `"'`)
		}
	}

	return ""
}

// 获取所有草稿文件
func getAllDrafts() []string {
	var drafts []string

	filepath.Walk(config.GetDraftDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			drafts = append(drafts, path)
		}

		return nil
	})

	return drafts
}

// 发布草稿到博客目录
func publishDraft(draftPath string) (string, error) {
	// 获取草稿目录的绝对路径
	absDraftDir, err := filepath.Abs(config.GetDraftDir())
	if err != nil {
		return "", fmt.Errorf("获取草稿目录绝对路径失败: %v", err)
	}

	// 确保输入路径是绝对路径
	absDraftPath, err := filepath.Abs(draftPath)
	if err != nil {
		return "", fmt.Errorf("获取草稿文件绝对路径失败: %v", err)
	}

	// 计算相对于草稿目录的路径
	relPath, err := filepath.Rel(absDraftDir, absDraftPath)
	if err != nil {
		return "", fmt.Errorf("计算相对路径失败: %v", err)
	}

	// 构建目标路径（保持相同的目录结构）
	targetPath := filepath.Join(config.GetBlogsDir(), relPath)

	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("目标文件已存在: %s", targetPath)
	}

	// 确保目标目录存在（只在需要时创建）
	targetDir := filepath.Dir(targetPath)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %v", err)
		}
	}

	// 读取原文件内容
	content, err := os.ReadFile(absDraftPath)
	if err != nil {
		return "", fmt.Errorf("读取草稿文件失败: %v", err)
	}

	// 更新文章末尾的时间戳
	updatedContent := updateTimestamp(string(content))

	// 写入目标文件
	if err := os.WriteFile(targetPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("写入目标文件失败: %v", err)
	}

	// 删除原草稿文件
	if err := os.Remove(absDraftPath); err != nil {
		// 如果删除失败，尝试删除已创建的目标文件
		os.Remove(targetPath)
		return "", fmt.Errorf("删除原草稿文件失败: %v", err)
	}

	return targetPath, nil
}

// 更新文章末尾的时间戳
func updateTimestamp(content string) string {
	now := time.Now()
	newTimestamp := fmt.Sprintf("> 更新时间: %s", now.Format("2006年01月02日 15:04"))

	// 使用正则表达式匹配并替换时间戳
	timestampRegex := regexp.MustCompile(`> 更新时间: .+`)

	if timestampRegex.MatchString(content) {
		// 如果找到现有时间戳，替换它
		return timestampRegex.ReplaceAllString(content, newTimestamp)
	} else {
		// 如果没有找到时间戳，在文章末尾添加
		content = strings.TrimRight(content, "\n")
		return content + "\n\n---\n\n" + newTimestamp + "\n"
	}
}
