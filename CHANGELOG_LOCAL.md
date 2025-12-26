# LGH v1.0.3 发布准备

**日期**: 2025-12-26

## ✅ 已完成的修改

### 1. `lgh status` 显示仓库列表
- **文件**: `cmd/lgh/status.go`
- **效果**: 在 "Repositories: N registered" 下方显示仓库名称列表

### 2. `lgh add` 自动初始化 Git 仓库
- **文件**: `cmd/lgh/add.go`, `internal/git/repo.go`
- **效果**: 对非 Git 目录自动执行 `git init`

### 3. 版本号更新 (1.0.2 → 1.0.3)
- `cmd/lgh/main.go`
- `Makefile`
- `README.md` (下载链接 x5)
- `README.zh-CN.md` (下载链接 x5)

### 4. 删除冗余的 Formula 目录
- **已删除**: `lgh/Formula/lgh.rb` (与 homebrew-tap 重复)

### 5. 添加对比表格到 README
- `README.md` - 英文版对比表
- `README.zh-CN.md` - 中文版对比表

### 6. 更新帮助文档
- `lgh add --help` - 说明自动初始化功能
- `lgh status --help` - 说明显示仓库列表

---

## 📦 homebrew-tap 项目更新
- **文件**: `Formula/lgh.rb`
- **状态**: 版本号已更新到 1.0.3
- **待办**: SHA256 需要等 release 发布后更新

---

## 🚀 发布步骤

### Step 1: 构建 release 二进制
```bash
cd /Users/fenge1222/neil/LocalGitHub/lgh
make release
```

### Step 2: 提交 lgh 项目
```bash
git add -A
git commit -m "v1.0.3: auto git-init, show repo list in status, add comparison table"
git tag v1.0.3
git push origin main --tags
```

### Step 3: 创建 GitHub Release
- 上传 `dist/` 目录下的二进制文件
- 记录每个文件的 SHA256 (在 `dist/checksums.txt`)

### Step 4: 更新 homebrew-tap
```bash
# 用实际的 SHA256 替换占位符
cd /Users/fenge1222/neil/LocalGitHub/homebrew-tap
# 编辑 Formula/lgh.rb 替换 PLACEHOLDER_XXX

git add -A
git commit -m "Update lgh to v1.0.3"
git push origin main
```

---

## 📝 v1.0.3 Release Notes (草稿)

### 🚀 新功能
- **自动 Git 初始化**: `lgh add .` 现在可以直接添加非 Git 目录，自动执行 `git init`
- **状态命令增强**: `lgh status` 现在会显示已注册仓库的名称列表，方便删除操作

### 📖 文档改进
- 添加与 GitLab、Gitea、git daemon 等方案的对比表格
- 更新命令帮助说明

### 🔧 维护
- 删除冗余的 Formula 目录（使用独立的 homebrew-tap 仓库）
