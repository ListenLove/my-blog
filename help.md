# MyBlog 使用帮助

## 快速开始

### 1. 构建项目
```bash
go build -o myblog.exe .
```

### 2. 查看和设置配置
```bash
# 查看当前配置
./myblog.exe config show

# 设置作者
./myblog.exe config set author "你的名字"

# 设置草稿目录
./myblog.exe config set directories.draft "my_drafts"

# 设置博客目录  
./myblog.exe config set directories.blogs "my_blogs"
```

### 3. 创建草稿文章

#### 方式一：带标签（推荐）
```bash
./myblog.exe draft "Go设计模式实践" --tags "Go,设计模式,实践"
# 将创建：_draft/Go/设计模式/实践/go设计模式实践.md
```

#### 方式二：不带标签
```bash
./myblog.exe draft "我的第一篇博客"
# 将创建：_draft/我的第一篇博客.md
```

#### 方式三：交互式模式（推荐）
```bash
./myblog.exe draft
# 交互式步骤：
# 1. 输入文章标题
# 2. 从现有标签中多选（可跳过）
# 3. 输入自定义标签（可跳过）
# 4. 自动合并并去重标签
```

### 4. 查看帮助
```bash
# 查看主帮助
./myblog.exe --help

# 查看draft命令帮助
./myblog.exe draft --help

# 查看config命令帮助
./myblog.exe config --help
```

## 配置管理

### 配置文件
MyBlog使用YAML格式的配置文件 `config.yaml`，包含以下配置项：

```yaml
# 作者信息
author: "你的名字"

# 目录配置
directories:
  # 草稿目录
  draft: "_draft"  
  # 博客目录
  blogs: "blogs"

# 文章默认配置
article:
  # 是否为草稿
  draft: true
  # 默认描述
  description: ""
```

### 配置命令

#### config show
显示当前配置信息，包括配置文件路径和所有设置项。

#### config set <key> <value>
设置配置项的值。

**可用配置项：**
- `author` - 作者名称
- `directories.draft` - 草稿目录路径
- `directories.blogs` - 博客目录路径
- `article.draft` - 文章默认草稿状态 (true/false)
- `article.description` - 文章默认描述

**示例：**
```bash
./myblog.exe config set author "张三"
./myblog.exe config set directories.draft "drafts"
./myblog.exe config set article.draft false
```

## 目录结构说明

### 标签即目录结构
- 标签 `"Go,设计模式"` → `_draft/Go/设计模式/`
- 标签 `"前端,Vue,组件"` → `_draft/前端/Vue/组件/`
- 无标签 → `_draft/` (根目录)

### 发布准备
这种目录结构设计是为了方便后续发布时：
- 可以按目录批量将草稿移动到 `blogs/` 目录
- 保持相同的目录结构便于管理
- 支持按标签分类浏览

## 命令说明

### draft 命令
创建一篇新的草稿文章，支持基于标签的目录结构。

**参数:**
- `title` - 文章标题（可选，如果不提供将进入交互式模式）

**选项:**
- `-t, --tags` - 文章标签，用逗号分隔，将作为目录结构（例如：Go,设计模式）
- `-v, --verbose` - 显示详细输出
- `-h, --help` - 显示帮助信息

## 交互式模式详解

交互式模式提供了最友好的用户体验，避免目录结构过于复杂：

### 步骤流程
1. **输入标题**: 输入文章标题（必须）
2. **选择标签路径**: 从现有完整路径中单选，或选择输入新路径
   - 现有路径如 `Go/设计模式`、`Java/Spring` 作为整体选项
   - 选择 `输入新标签路径` 来创建全新的目录结构
3. **预览结构**: 显示最终的目录结构预览

### 设计理念
- **保持整洁**: 不拆分现有路径，避免目录结构混乱
- **单一选择**: 每次只选择一个完整的标签路径
- **简化管理**: 现有的 `Go/设计模式` 不会拆分为单独的 `Go` 和 `设计模式`

### 使用示例
```
$ ./myblog.exe draft

请输入文章标题: 单例模式实现

请选择标签路径:
> 输入新标签路径
  Go/设计模式/实践  
  Java/Spring/Boot
  测试/单选/功能

# 如果选择现有路径：
📁 目录结构预览: Go,设计模式,实践 → _draft/Go/设计模式/实践/

# 如果选择输入新路径：
请输入标签路径 (用逗号分隔创建多级目录): Go,设计模式,单例

📁 目录结构预览: Go,设计模式,单例 → _draft/Go/设计模式/单例/
```

## 使用示例

```bash
# 1. 创建Go相关文章
./myblog.exe draft "Go并发编程" --tags "Go,并发,编程"
# → _draft/Go/并发/编程/go并发编程.md

# 2. 创建前端文章  
./myblog.exe draft "Vue组件开发" --tags "前端,Vue,组件"
# → _draft/前端/Vue/组件/vue组件开发.md

# 3. 创建算法文章
./myblog.exe draft "二分查找算法" --tags "算法,查找,二分"
# → _draft/算法/查找/二分/二分查找算法.md

# 4. 无标签文章（直接放在根目录）
./myblog.exe draft "个人随笔"
# → _draft/个人随笔.md

```bash
# 5. 交互式创建（推荐）
./myblog.exe draft
# 单选现有完整路径或输入新路径，保持目录结构整洁
```

## 标签路径管理

### 现有路径保持完整
- 现有 `Go/设计模式/实践` 作为一个整体选项
- 现有 `Java/Spring/Boot` 作为一个整体选项  
- 不会拆分为单独的标签供自由组合

### 创建新路径
- 用户可以输入 `Go,并发,Goroutine` 创建新路径
- 系统会创建 `_draft/Go/并发/Goroutine/` 目录结构
- 新路径会在后续作为完整选项出现

### 优势
- **避免混乱**: 防止标签自由组合产生过多目录
- **保持整洁**: 现有结构不会被拆散重组
- **便于管理**: 每个路径代表一个明确的分类体系

## 生成的文章格式

每篇文章都会自动生成包含以下信息的Front Matter：

```yaml
---
title: "文章标题"
date: 2025-09-01T15:04:05Z07:00
draft: true
tags: ["Go", "设计模式", "实践"]
author: "MyBlog"
description: ""
---
```

文章末尾会包含创建信息和标签路径：
```markdown
---

> 这是一篇草稿文章，创建于 2025年09月01日 17:29  
> 使用 MyBlog 静态博客系统生成  
> 标签路径: Go/设计模式/实践
```

## 项目结构

```
my-blog/
├── _draft/              # 草稿文章目录（按标签组织）
│   ├── Go/
│   │   ├── 设计模式/
│   │   │   └── 实践/
│   │   │       └── go设计模式实践.md
│   │   └── 并发/
│   ├── 前端/
│   │   └── Vue/
│   └── 无标签文章.md
├── blogs/               # 正式文章目录（用于获取已有标签）
├── cmd/                 # 命令目录
│   └── draft.go         # Draft命令实现
├── main.go             # 主程序入口
├── go.mod              # Go模块定义
├── README.md           # 项目说明
├── help.md             # 使用帮助（本文件）
└── myblog.exe          # 编译后的可执行文件
```

## 智能标签提示

交互式模式会扫描 `blogs/` 和 `_draft/` 目录，提取已有的标签供参考：

```
请输入标签 (用逗号分隔，将作为目录结构):
例如: Go,设计模式 -> _draft/Go/设计模式/
已有标签: Go, 设计模式, 实践, 前端, Vue, 组件
```

## 常见问题

**Q: 为什么要用标签作为目录结构？**
A: 这样设计便于后续发布时按目录批量操作，同时保持文章的分类清晰。

**Q: 如何修改生成的文章模板？**
A: 编辑 `cmd/draft.go` 文件中的 `generateMarkdownContent` 函数。

**Q: 标签名称有限制吗？**
A: 避免使用文件系统不支持的字符，程序会自动处理特殊字符。

**Q: 可以创建嵌套很深的目录吗？**
A: 可以，但建议标签层级不超过3-4层，便于管理。
