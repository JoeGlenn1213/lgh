# lgh Evolution Notes

> 下一阶段演化方向探讨，非执行计划

---

## 1. 当前定位回顾

lgh 目前是：
- 本地 Git 服务器 (bare repo)
- mDNS 局域网发现
- HMAC-SHA256 认证
- MCP Server (AI 可控)
- Git 事件源 → ActionD

**核心原则**: lgh 只做 Git 服务，不做业务逻辑

---

## 2. 关键决策点

### 决策 2.1: 是否引入插件机制？

**现状**: lgh 是单体，无插件扩展

**选项**:

| 选项 | 优点 | 缺点 |
|------|------|------|
| A. 保持单体 | 简单、安全 | 扩展需改源码 |
| B. 引入插件 | 可扩展 | 增加复杂度 |

**建议**: 保持单体，通过 MCP 接口与外部交互

---

### 决策 2.2: 是否支持 Webhook / Event ？

**现状**: mDNS 推送事件给 ActionD

**选项**:

| 选项 | 优点 | 缺点 |
|------|------|------|
| A. 保持现状 | 简单、本地优先 | 跨网络困难 |
| B. 支持 Webhook | 跨网络、跨服务 | 复杂度增加 |

**建议**: 未来考虑可选 webhook，但不是 V1 优先级

---

### 决策 2.3: 是否成为 ActionD 正式事件源？

**现状**: lgh 通过 socket 推送事件，非正式协议

**建议**: 
- 定义正式 Event 协议
- lgh 成为 "Event Source" 而非直接推送
- ActionD 消费事件，保持解耦

---

### 决策 2.4: 是否抽离 server 层？

**现状**: cmd/lgh 和 internal/server 耦合较紧

**建议**:
- 保持当前结构
- 如果未来需要微服务拆分，再考虑分层
- 当前优先级：稳定性 > 重构

---

## 3. 架构预留设计

### 3.1 未来可能的模块拆分

```
lgh/
├── cmd/lgh          # CLI 入口
├── cmd/lgh-server   # 可选的独立 server 进程
├── internal/
│   ├── git/         # Git 操作封装
│   ├── auth/        # 认证逻辑
│   ├── registry/    # 仓库注册
│   ├── event/       # 事件推送（可抽象为接口）
│   └── mcp/         # MCP Server
└── pkg/
    └── skill/       # Skills 机制
```

### 3.2 Event 接口抽象

```go
// EventPublisher is the interface for publishing git events
type EventPublisher interface {
    Publish(event GitEvent) error
    Subscribe(handler EventHandler)
}

// Concrete implementations:
// - MDNSPublisher (current)
// - WebhookPublisher (future)
// - SocketPublisher (current, for ActionD)
```

### 3.3 插件化预留（可选）

```go
// PreActionPlugin runs before git operations
type PreActionPlugin interface {
    Name() string
    PreHook(ctx *HookContext) error
}

// PostActionPlugin runs after git operations  
type PostActionPlugin interface {
    Name() string
    PostHook(ctx *HookContext) error
}
```

---

## 4. 与 ActionD 的事件边界

### 当前状态
```
lgh (push) → socket → ActionD
```

### 建议的正式协议

```go
// GitEvent is the canonical event format
type GitEvent struct {
    Type      string    // git.push, git.tag
    Repo      string    // repository name
    Branch    string    // branch name (for push)
    Tag       string    // tag name (for tag)
    CommitSHA string    // commit hash
    Committer string    // who triggered
    Timestamp time.Time // when
    
    // Metadata for routing
    Source    string    // "lgh" or "external"
    Profile   string    // optional: fast, full, release
}
```

---

## 5. 下一步行动（未来）

| 优先级 | 行动 | 说明 |
|--------|------|------|
| P2 | 定义 Event 协议 | 文档化 git push/tag 事件格式 |
| P3 | 抽离 Event 模块 | 便于测试和 Mock |
| P4 | 可选 Webhook | 跨网络支持 |

---

## 6. 当前约束

**不要做**:
- 不要引入数据库（SQLite 是 ActionD 的职责）
- 不要做 CI/CD 逻辑（ActionD 负责）
- 不要做 Web UI（actiond-web 负责）
- 不要过度设计（保持简单）

**坚持做**:
- Git 服务稳定性
- 认证安全
- 事件可靠性
- MCP 接口完整

---

## 7. 总结

lgh 的演化方向：**保持专注，成为最好的本地 Git 服务器**

- 核心不变：Git bare repo + mDNS + MCP
- 逐步开放：Event 协议抽象
- 不做：业务逻辑、Web UI、复杂插件
