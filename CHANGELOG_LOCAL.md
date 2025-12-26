# LGH v1.0.4 发布准备

**日期**: 2025-12-26

## ✅ 新功能 (v1.0.4)

### 1. 核心仓库状态工具 (`lgh repo`)
- **lgh repo status**: 在任何 git 项目目录中，清晰展示本地与远程的连接状态
- **lgh repo inspect**: 查看 LGH 内部裸仓的详细信息 (HEAD, 分支, 最近提交)
- **lgh repo set-default**: 修改裸仓的默认分支 (HEAD symbolic-ref)

### 2. 远程切换器 (`lgh remote`)
- **lgh remote use**: 快速切换当前分支的上游 (upstream)，例如在 `lgh` 和 `origin` 之间切换

### 3. 便捷工具
- **lgh clone**: 语法糖，`lgh clone my-repo` 直接克隆本地仓库
- **lgh doctor**: 系统健康检查，检测环境、配置和端口问题

## 📝 变更文件
- `cmd/lgh/repo.go` (新增)
- `cmd/lgh/remote.go` (新增)
- `cmd/lgh/clone.go` (新增)
- `cmd/lgh/doctor.go` (新增)
- `cmd/lgh/main.go` (注册新命令，更新版本号)
- `internal/git/repo.go` (增强 git 功能支持)
- `pkg/ui/output.go` (增强 UI 支持)
- `Makefile` (版本号 1.0.4)
- `README.md` / `README.zh-CN.md` (文档更新)

---

## 🚀 发布步骤

### Step 1: 构建 release 二进制
```bash
make release
```

### Step 2: 提交代码
```bash
git add -A
git commit -m "v1.0.4: Add repo status/inspect, remote switcher, doctor, and clone commands"
git tag v1.0.4
git push origin main --tags
```

### Step 3: 创建 GitHub Release
- Tag: `v1.0.4`
- Title: `LGH v1.0.4 - The Repository Management Update`
- Upload binaries
- Copy SHA256

### Step 4: 更新 homebrew-tap
- 更新 `Formula/lgh.rb` 中的 URL 和 SHA256

---
