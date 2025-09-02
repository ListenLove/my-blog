package cmd

import (
	"MyBlog/internal/config"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	genVerbose bool
)

var GenCmd = &cobra.Command{
	Use:   "gen",
	Short: "生成README.md文档，按标签分类展示所有已发布的文章",
	Long: `生成README.md文档，将已发布的文章按标签分门别类地整理展示。

该命令会：
1. 扫描blogs目录中的所有文章
2. 解析文章的Front Matter获取标题和标签信息
3. 按标签分类整理所有文章
4. 生成包含快速导航和文章分类的README.md文档`,
	Example: `  myblog gen
  myblog gen --verbose`,
	Args: cobra.NoArgs,
	Run:  runGenCommand,
}

func init() {
	// 添加命令行标志
	GenCmd.Flags().BoolVarP(&genVerbose, "verbose", "v", false, "详细输出")

	// 设置日志级别
	if genVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

type GenArticleInfo struct {
	Title        string    `yaml:"title"`
	Date         time.Time `yaml:"date"`
	Published    time.Time `yaml:"published"`
	Tags         []string  `yaml:"tags"`
	FilePath     string    `yaml:"-"`
	RelativePath string    `yaml:"-"`
}

type TagGroup struct {
	TagPath  string
	Articles []GenArticleInfo
}

func runGenCommand(cmd *cobra.Command, args []string) {
	// 设置颜色输出
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	fmt.Printf("%s 开始扫描已发布的文章...\n", blue("信息:"))

	// 扫描blogs目录获取所有文章
	articles, err := scanPublishedArticles()
	if err != nil {
		fmt.Printf("%s 扫描文章失败: %v\n", red("错误:"), err)
		logrus.WithError(err).Error("扫描文章失败")
		return
	}

	if len(articles) == 0 {
		fmt.Printf("%s 没有找到已发布的文章\n", yellow("提示:"))
		return
	}

	fmt.Printf("%s 找到 %d 篇已发布的文章\n", blue("信息:"), len(articles))

	// 按标签分组文章
	tagGroups := groupArticlesByTags(articles)

	fmt.Printf("%s 按标签分组完成，共 %d 个标签分类\n", blue("信息:"), len(tagGroups))

	// 生成README.md
	err = generateReadme(tagGroups)
	if err != nil {
		fmt.Printf("%s 生成README.md失败: %v\n", red("错误:"), err)
		logrus.WithError(err).Error("生成README.md失败")
		return
	}

	fmt.Printf("%s 成功生成README.md文档!\n", green("✓"))
	fmt.Printf("  文章总数: %s\n", yellow(fmt.Sprintf("%d", len(articles))))
	fmt.Printf("  标签分类: %s\n", yellow(fmt.Sprintf("%d", len(tagGroups))))
	fmt.Printf("  文件路径: %s\n", green("README.md"))

	logrus.WithFields(logrus.Fields{
		"articles_count": len(articles),
		"tag_groups":     len(tagGroups),
	}).Info("README.md生成成功")
}

func scanPublishedArticles() ([]GenArticleInfo, error) {
	var articles []GenArticleInfo
	blogsDir := config.GetBlogsDir()

	// 检查blogs目录是否存在
	if _, err := os.Stat(blogsDir); os.IsNotExist(err) {
		// 目录不存在，返回空切片而不是错误
		return articles, nil
	}

	err := filepath.Walk(blogsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.md文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			article, err := parseArticle(path)
			if err != nil {
				logrus.WithError(err).Warnf("解析文章失败: %s", path)
				return nil // 继续处理其他文件，不中断整个过程
			}
			
			// 计算相对路径
			relPath, err := filepath.Rel(blogsDir, path)
			if err != nil {
				relPath = path
			}
			article.RelativePath = filepath.ToSlash(relPath)
			
			articles = append(articles, *article)
		}

		return nil
	})

	return articles, err
}

func parseArticle(filePath string) (*GenArticleInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var frontMatterLines []string
	inFrontMatter := false
	frontMatterEnd := false

	for scanner.Scan() {
		line := scanner.Text()
		
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				frontMatterEnd = true
				break
			}
		}

		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	if !frontMatterEnd {
		return nil, fmt.Errorf("找不到完整的Front Matter")
	}

	// 解析YAML Front Matter
	frontMatterContent := strings.Join(frontMatterLines, "\n")
	
	// 使用 viper 解析 YAML
	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(strings.NewReader(frontMatterContent))
	if err != nil {
		return nil, fmt.Errorf("解析Front Matter失败: %v", err)
	}

	var article GenArticleInfo
	article.Title = v.GetString("title")
	article.Tags = v.GetStringSlice("tags")
	
	// 直接从viper获取时间
	article.Published = v.GetTime("published")
	article.Date = v.GetTime("date")
	
	// 如果published时间为空，使用date时间
	if article.Published.IsZero() && !article.Date.IsZero() {
		article.Published = article.Date
	}


	article.FilePath = filePath
	
	// 如果标题为空，使用文件名作为标题
	if article.Title == "" {
		filename := filepath.Base(filePath)
		article.Title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	return &article, nil
}

func groupArticlesByTags(articles []GenArticleInfo) []TagGroup {
	tagMap := make(map[string][]GenArticleInfo)

	for _, article := range articles {
		if len(article.Tags) == 0 {
			// 没有标签的文章放到"其他"分类
			tagPath := "其他"
			tagMap[tagPath] = append(tagMap[tagPath], article)
		} else {
			// 构建标签路径
			tagPath := strings.Join(article.Tags, "/")
			tagMap[tagPath] = append(tagMap[tagPath], article)
		}
	}

	// 转换为切片并排序
	var tagGroups []TagGroup
	for tagPath, articles := range tagMap {
		// 按发布时间排序文章（最新的在前）
		sort.Slice(articles, func(i, j int) bool {
			return articles[i].Published.After(articles[j].Published)
		})
		
		tagGroups = append(tagGroups, TagGroup{
			TagPath:  tagPath,
			Articles: articles,
		})
	}

	// 按标签路径排序
	sort.Slice(tagGroups, func(i, j int) bool {
		return tagGroups[i].TagPath < tagGroups[j].TagPath
	})

	return tagGroups
}

func generateReadme(tagGroups []TagGroup) error {
	var content strings.Builder

	// 写入项目介绍
	content.WriteString(`# MyBlog - 简易静态博客系统

MyBlog 是一个简易的静态博客系统，类似于 Hugo，专为命令行使用而设计。它借助 GitHub 对 Markdown 文档的完美支持，让 Markdown 文档成为你的界面。

## 特性

- 🚀 **简单易用**: 基于命令行的简洁界面
- 📝 **Markdown 支持**: 完美支持 Markdown 格式
- 🏷️ **标签和分类**: 支持文章标签和分类管理
- 🎨 **交互式命令**: 美观的交互式命令行界面
- 🌈 **彩色输出**: 支持彩色终端输出
- 📁 **自动组织**: 自动创建和组织文件结构

## 快速开始

### 构建项目

`)
	content.WriteString("```bash\n")
	content.WriteString("go build -o myblog .\n")
	content.WriteString("```\n\n")

	// 添加帮助文档链接
	content.WriteString(`### 使用帮助

详细使用说明请参考：[使用帮助文档](help.md)

`)

	// 生成文章标签快速导航
	content.WriteString("## 📚 文章导航\n\n")
	
	// 生成目录
	content.WriteString("### 标签分类快速跳转\n\n")
	for _, group := range tagGroups {
		anchorLink := strings.ToLower(strings.ReplaceAll(group.TagPath, "/", "-"))
		anchorLink = strings.ReplaceAll(anchorLink, " ", "-")
		content.WriteString(fmt.Sprintf("- [%s](#%s) (%d篇)\n", group.TagPath, anchorLink, len(group.Articles)))
	}
	content.WriteString("\n")

	// 统计信息
	totalArticles := 0
	for _, group := range tagGroups {
		totalArticles += len(group.Articles)
	}
	
	content.WriteString(fmt.Sprintf("**📊 统计信息**: 共 %d 个标签分类，%d 篇文章\n\n", len(tagGroups), totalArticles))
	content.WriteString("---\n\n")

	// 按标签分类展示文章
	content.WriteString("## 📖 文章分类\n\n")

	for _, group := range tagGroups {
		// 标签分类标题
		content.WriteString(fmt.Sprintf("### %s\n\n", group.TagPath))
		
		// 文章列表
		for _, article := range group.Articles {
			publishedTime := article.Published.Format("2006-01-02")
			
			// 生成GitHub相对路径链接
			articleLink := fmt.Sprintf("blogs/%s", article.RelativePath)
			
			content.WriteString(fmt.Sprintf("- [%s](%s) - *%s*\n", 
				article.Title, 
				articleLink, 
				publishedTime))
		}
		
		content.WriteString("\n")
	}

	// 添加项目信息
	content.WriteString("---\n\n")
	content.WriteString("## 可用命令\n\n")
	content.WriteString("- `draft` - 创建草稿文章\n")
	content.WriteString("- `new` - 创建正式文章\n")
	content.WriteString("- `pub` - 发布草稿到正式文章\n")
	content.WriteString("- `gen` - 生成README.md文档\n\n")
	
	// 生成时间
	content.WriteString(fmt.Sprintf("*README.md 生成时间: %s*\n", time.Now().Format("2006-01-02 15:04:05")))

	// 写入文件
	return os.WriteFile("README.md", []byte(content.String()), 0644)
}
