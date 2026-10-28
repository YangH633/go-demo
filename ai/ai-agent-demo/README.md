## ai-agent-demo

一个用于演示在 Go 中集成 LLM / Agent 能力的示例仓库，包含：

- **framework_agent**: 使用 `langchaingo` 框架的简单 Agent Demo（当前示例）
下面主要介绍 `framework_agent` 这个子示例的用法。

---

### 1. 运行环境

- **Go**: 推荐 Go 1.22+（参考仓库根目录 `go.mod`）
- **依赖管理**: 使用 Go Modules
- **LLM 提供方**: 兼容 OpenAI 接口的服务（示例以 DeepSeek 为主）

你可以使用本地安装的 Go 环境，或使用 `asdf` / `mise` 等工具（`ai-agent-demo/.tool-versions` 中给出了参考版本）。

---

### 2. 环境变量配置

在 `ai-agent-demo` 目录下有一个 `.env.example`，复制一份为 `.env` 并根据实际情况填写：

```bash
cp .env.example .env
```

常用变量说明（任选其一提供 API Key 即可）：

- **DEEPSEEK_API_KEY**: DeepSeek 的 API Key
- **OPENAI_API_KEY**: 兼容 OpenAI 接口的 API Key（如果你直接用 OpenAI）
- **DEEPSEEK_MODEL**: 模型名称，默认 `deepseek-chat`
- **DEEPSEEK_BASE_URL**: 接口地址，默认 `https://api.deepseek.com/v1`

`framework_agent/main.go` 会优先读取：

1. `DEEPSEEK_API_KEY`
2. 若为空，则读取 `OPENAI_API_KEY`
3. 两者都缺失时会直接报错退出

---

### 3. 安装依赖

在仓库根目录（`go-demo`）或 `ai-agent-demo` 目录执行：

```bash
go mod tidy
```

若只想为 `framework_agent` 子模块整理依赖，可以进入该目录执行：

```bash
cd framework_agent
go mod tidy
```

---

### 4. 运行 framework_agent Demo

进入 `ai-agent-demo` 根目录，执行：

```bash
cd ai-agent-demo
go run ./framework_agent
```

程序会：

1. 加载 `.env` 中的环境变量
2. 使用 `langchaingo` 创建一个基于 DeepSeek/OpenAI 的 LLM 客户端
3. 注册一个内置的 `Calculator` 工具
4. 构建一个 `OneShotAgent`，让它自己决定是否使用计算器
5. 询问一个关于「1.05 的 30 次方」的复利类比问题，并输出最终回答

终端中你会看到：

- Agent 提出的问题内容
- 它在调用 `Calculator` 工具时的中间思考过程（由 `langchaingo` 决定输出格式）
- 最终的中文回答结果

---

### 5. 修改成你自己的 Agent Demo

你可以直接修改 `framework_agent/main.go` 中的内容，快速做一些试验：

- **修改提问内容**: 变更 `question := "..."` 为你想让 Agent 解决的问题
- **调整推理轮数**: 调整 `agents.WithMaxIterations(3)` 中的数字
- **扩展工具列表**: 在 `agentTools := []tools.Tool{ ... }` 里增加你自定义的工具

这份 demo 的定位是「最小可运行的 Agent 示例」，方便你在本地快速确认：

- 环境变量 / API Key 是否配置正确
- 本地到 LLM 服务的网络是否畅通
- `langchaingo` 的基本用法是否理解