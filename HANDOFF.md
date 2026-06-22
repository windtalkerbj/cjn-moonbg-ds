# Moon Bridge 修改记录 — 2026-06-22

## 背景

Codex 配置了 `model_reasoning_effort = "xhigh"` + `model = "deepseek-v4-pro"`，但 thinking_mode 实际未生效。排查发现 Moon Bridge 发给 DeepSeek 的 Anthropic 请求中缺少 `output_config` 字段。

## 问题根因

Moon Bridge 的 OpenAIClient → CoreRequest → AnthropicRequest 适配器链路中：

- `deepseek_v4` 插件的 `MutateRequest` 钩子（旧路径）能正确设置 `req.Output.Effort`，但新适配器路径不调用它
- 新路径只调用 `MutateCoreRequest`，而该函数只设置了 `Extensions["thinking"]["budget_tokens"]`，未从 `Extensions["openai"]["reasoning"]["effort"]` 中提取 effort 并设置 `req.Output`

## 修改内容

### 1. `config.yml` — 开启 trace + debug 日志

在配置文件顶部新增：

```yaml
log:
  level: "debug"
  format: "text"

trace:
  enabled: true
```

trace 输出路径：`data/trace/Transform/<timestamp>/`

### 2. `internal/extension/deepseek_v4/plugin.go` — 修复 MutateCoreRequest

在 `MutateCoreRequest` 中，`req.Extensions` 初始化之后、budget_tokens 读取之前，新增 reasoning effort 提取逻辑：

```go
// Extract reasoning effort from the OpenAI request extension.
// Codex sends reasoning.effort (xhigh/high) which must be mapped to
// DeepSeek's Anthropic-compatible output_config.effort (max/high).
if req.Output == nil {
    if openaiExt, ok := req.Extensions["openai"].(map[string]any); ok {
        if reasoning, ok := openaiExt["reasoning"].(map[string]any); ok {
            if effort, yes := reasoningEffort(reasoning); yes {
                req.Output = &format.CoreOutputConfig{Effort: effort}
            }
        }
    }
}
```

`reasoningEffort` 函数已存在于同文件，负责映射 `xhigh` → `"max"`、`high` → `"high"`。

### 3. 编译 & 部署

```bash
cd /Users/cjn/moon-bridge
go build -o moonbridge ./cmd/...
# 重启服务
```

## 验证结果

修复后 trace 验证：
- Anthropic 请求包含 `output_config: {"effort": "max"}`
- 流事件包含 30+ `thinking_delta` 事件
- Response 侧 Output[0] 为 `type=reasoning`，含完整思考文本

## 文件清单

| 文件 | 修改 |
|------|------|
| `config.yml` | 新增 `log.debug` + `trace.enabled` |
| `internal/extension/deepseek_v4/plugin.go` | `MutateCoreRequest` 新增 8 行 effort 提取 |
| `moonbridge` (binary) | 重新编译 |
