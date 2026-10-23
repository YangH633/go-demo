package main

import (
	"context"
	"fmt"
	"os"

	"ai-agent-demo/env"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

// 这是一个使用 LangChainGo(langchaingo) 的简单 Agent Demo。
// 它使用 OpenAI 模型 + 内置 Calculator 工具，让 Agent 可以自己决定是否调用计算器。

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "framework agent error:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := env.LoadEnv(); err != nil {
		return fmt.Errorf("加载环境变量失败: %w", err)
	}

	// DeepSeek 提供 OpenAI 兼容接口：配置 BaseURL + Model + Token 即可复用 openai client。
	// 你可以把 DEEPSEEK_API_KEY 写到 .env；或直接复用 OPENAI_API_KEY。
	token := os.Getenv("DEEPSEEK_API_KEY")
	if token == "" {
		token = os.Getenv("OPENAI_API_KEY")
	}
	if token == "" {
		return fmt.Errorf("缺少 API Key：请设置 DEEPSEEK_API_KEY（或 OPENAI_API_KEY）")
	}

	// DeepSeek 常用模型名：deepseek-chat / deepseek-reasoner
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	// DeepSeek 的 OpenAI 兼容 Base URL（通常是 https://api.deepseek.com/v1）
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	llm, err := openai.New(
		openai.WithToken(token),
		openai.WithModel(model),
		openai.WithBaseURL(baseURL),
		openai.WithAPIType(openai.APITypeOpenAI),
	)
	if err != nil {
		return fmt.Errorf("创建 openai LLM 失败: %w", err)
	}

	// 准备工具，这里只用内置计算器，避免额外 API Key 依赖。
	agentTools := []tools.Tool{
		tools.Calculator{},
	}

	// 创建一个 OneShotAgent：每个问题最多若干轮思考 + 调工具。
	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(3),
	)

	// Executor 负责把“自然语言问题 -> Agent -> 工具 -> 最终答案”这一整条链跑完。
	executor := agents.NewExecutor(agent)

	question := "帮我估算一下，1.05 的 30 次方大约是多少，并简单解释这个数字代表什么含义（可以用复利的角度类比）。"

	fmt.Println("问题：", question)
	fmt.Println("思考中（可能会多次调用工具 Calculator）...")
	fmt.Println()

	answer, err := chains.Run(context.Background(), executor, question)
	if err != nil {
		return fmt.Errorf("运行 agent executor 失败: %w", err)
	}

	fmt.Println("Agent 最终回答：")
	fmt.Println(answer)
	return nil
}
