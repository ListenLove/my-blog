# MyBlog - 简易静态博客系统

MyBlog 是一个简易的静态博客系统，类似于 Hugo，专为命令行使用而设计。它借助 GitHub 对 Markdown 文档的完美支持，让 Markdown 文档成为你的界面。

## 特性

- 🚀 **简单易用**: 基于命令行的简洁界面
- 📝 **Markdown 支持**: 完美支持 Markdown 格式
- 🏷️ **标签和分类**: 支持文章标签和分类管理
- 🎨 **交互式命令**: 美观的交互式命令行界面
- 🌈 **彩色输出**: 支持彩色终端输出
- 📁 **自动组织**: 自动创建和组织文件结构

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 构建项目

```bash
go build -o myblog.exe .
```

### 配置管理

```bash
# 查看当前配置
./myblog.exe config show

# 设置作者
./myblog.exe config set author "你的名字"

# 设置目录
./myblog.exe config set directories.draft "my_drafts"
./myblog.exe config set directories.blogs "my_posts"
```

### 基本用法

#### 1. 创建草稿文章

```bash
# 方式1: 带标签创建（推荐）
./myblog.exe draft "Go设计模式实践" --tags "Go,设计模式,实践"

# 方式2: 不带标签创建
./myblog.exe draft "我的第一篇博客"

# 方式3: 交互式模式
./myblog.exe draft
```

#### 2. 查看帮助

```bash
# 查看主帮助
./myblog.exe --help

# 查看特定命令帮助
./myblog.exe draft --help
```

## 可用命令

### `draft` 命令

创建一篇新的草稿文章，支持基于标签的目录结构。

### `config` 命令

管理配置文件，支持查看和修改配置项。

**子命令:**
- `config show` - 显示当前配置
- `config set <key> <value>` - 设置配置项

**语法:**
```bash
myblog draft [title] [flags]
```

**标志:**
- `-t, --tags strings`: 文章标签（逗号分隔，将作为目录结构）
- `-v, --verbose`: 详细输出
- `-h, --help`: 显示帮助信息

**目录结构:**
- 标签 `"Go,设计模式"` → `_draft/Go/设计模式/`
- 无标签 → `_draft/` (根目录)

**示例:**
```bash
# 基本用法（带标签）
./myblog.exe draft "Go并发编程" --tags "Go,并发,编程"

# 不带标签
./myblog.exe draft "个人随笔"

# 交互式模式
./myblog.exe draft
```

## 项目结构

```
my-blog/
├── _draft/                 # 草稿文章目录（按标签组织）
│   ├── Go/
│   │   ├── 设计模式/
│   │   │   └── 实践/
│   │   │       └── go设计模式实践.md
│   │   └── 并发/
│   ├── 前端/
│   │   └── Vue/
│   └── 无标签文章.md
├── blogs/                  # 正式文章目录（用于获取已有标签）
├── cmd/                    # 命令目录
├── internal/               # 内部包
├── main.go                 # 主入口文件
├── draft.go                # Draft 命令实现
├── go.mod                  # Go 模块文件
├── help.md                 # 使用帮助文档
└── README.md              # 说明文档
```

## 生成的文章格式

每篇文章都会包含以下 Front Matter：

```yaml
---
title: "文章标题"
date: 2025-09-01T15:04:05Z07:00
draft: true
tags: ["tag1", "tag2"]
categories: ["分类"]
author: "MyBlog"
description: ""
---
```

## 使用的开源库

本项目使用了以下优秀的 Go 开源库：

- **[Cobra](https://github.com/spf13/cobra)**: 强大的命令行框架
- **[Survey](https://github.com/AlecAivazis/survey)**: 优雅的交互式命令行工具
- **[Logrus](https://github.com/sirupsen/logrus)**: 结构化日志库
- **[Color](https://github.com/fatih/color)**: 彩色终端输出

## 开发计划

- [ ] 添加 `publish` 命令将草稿转为正式文章
- [ ] 添加 `list` 命令查看所有文章
- [ ] 添加 `delete` 命令删除文章
- [ ] 添加 `serve` 命令本地预览
- [ ] 支持配置文件
- [ ] 支持模板自定义
- [ ] 添加文章搜索功能

## 贡献

欢迎提交 Issues 和 Pull Requests！

## 许可证

MIT License
