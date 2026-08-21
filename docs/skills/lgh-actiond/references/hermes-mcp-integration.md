# Hermes MCP Integration — LGH + ActionD

> 验证日期: 2026-08-20
> LGH: v1.3.0 (2026-08-13) / ActionD: v1.2.1 (2026-08-20)

## 配置方法

在 `~/.hermes/config.yaml` 的 `mcp_servers:` 段落添加：

```yaml
mcp_servers:
  lgh:
    command: /usr/local/bin/lgh
    args: [mcp]
    enabled: true
  actiond:
    command: /Users/fenge1222/.local/bin/actiond
    args: [mcp]
    enabled: true
```

> ⚠️ `hermes mcp add` 交互式提示会阻塞（需要 y/N 确认），建议直接编辑 config.yaml。

## 验证

```bash
hermes mcp list       # 确认 lgh / actiond 状态为 ✓ enabled
hermes mcp test lgh     # 应发现 10 个工具
hermes mcp test actiond # 应发现 20 个工具
```

## 生效方式

- 新 session 自动加载
- 当前 session: `/reload-mcp`（会 invalidate prompt cache，需确认）
- 或重启 session

## 工具清单（2026-08-20 验证）

### LGH (10 tools)
lgh_add, lgh_list, lgh_log, lgh_remove, lgh_rollback, lgh_save,
lgh_serve_start, lgh_serve_stop, lgh_status, lgh_up

### ActionD (20 tools)
actiond_action_get, actiond_actions_list, actiond_cancel, actiond_diagnose,
actiond_job_cancel, actiond_job_retry, actiond_job_wait, actiond_log,
actiond_plugin_disable, actiond_plugin_enable, actiond_plugins_list,
actiond_plugins_recommend, actiond_plugins_reload, actiond_profile_get,
actiond_profile_set, actiond_server_restart, actiond_server_start,
actiond_server_stop, actiond_status, dev_cycle_run

## 路径

| 组件 | 安装路径 |
|------|----------|
| LGH | `/usr/local/bin/lgh` |
| ActionD | `~/.local/bin/actiond` |
