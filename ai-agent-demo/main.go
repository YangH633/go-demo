package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-agent-demo/env"
)

// Message represents a single chat message to/from the model.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest matches the OpenAI /v1/chat/completions API.
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatCompletionResponse is a minimal struct for parsing responses.
type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// ToolCall is a simple JSON protocol for tool usage.
// Example: {"tool":"get_time","args":{}}
type ToolCall struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

// Agent holds conversation state and API config.
type Agent struct {
	SystemPrompt string
	History      []Message
	Model        string
	APIKey       string
	APIURL       string
}

// NewAgent constructs a new Agent instance.
func NewAgent(systemPrompt string) *Agent {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("请先设置环境变量 OPENAI_API_KEY")
		os.Exit(1)
	}

	return &Agent{
		SystemPrompt: systemPrompt,
		History:      []Message{},
		Model:        "deepseek-chat", // TODO: 替换为你实际可用的模型名称
		APIKey:       apiKey,
		APIURL:       "https://api.deepseek.com/chat/completions",
	}
}

// Run starts a simple REPL loop.
func (a *Agent) Run(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Go AI Agent Demo 已启动，输入内容开始对话，输入 `exit` 退出。")

	for {
		fmt.Print("\n你：")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入错误:", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println("Agent 已退出。")
			return
		}

		// Append user message.
		a.History = append(a.History, Message{
			Role:    "user",
			Content: line,
		})

		// Call model.
		reply, err := a.CallLLM(ctx)
		if err != nil {
			fmt.Println("调用 LLM 失败:", err)
			continue
		}

		// Try to interpret reply as a tool call.
		toolReply, ok := a.handleToolIfNeeded(reply)
		if ok {
			fmt.Println("Agent（工具调用）：", toolReply)
			// 也把工具结果记入历史，方便后续对话参考。
			a.History = append(a.History, Message{
				Role:    "assistant",
				Content: toolReply,
			})
			continue
		}

		// Normal assistant reply.
		a.History = append(a.History, Message{
			Role:    "assistant",
			Content: reply,
		})

		fmt.Println("Agent：", reply)
	}
}

// handleToolIfNeeded tries to parse the model reply as a ToolCall JSON and execute it.
// Returns (response, true) if a tool was executed, otherwise ("", false).
func (a *Agent) handleToolIfNeeded(reply string) (string, bool) {
	trimmed := strings.TrimSpace(reply)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return "", false
	}

	var tc ToolCall
	if err := json.Unmarshal([]byte(trimmed), &tc); err != nil || tc.Tool == "" {
		return "", false
	}

	switch tc.Tool {
	case "get_time":
		now := time.Now().Format("2006-01-02 15:04:05")
		return fmt.Sprintf("当前本地时间是：%s", now), true
	case "read_file":
		pathVal, ok := tc.Args["path"].(string)
		if !ok || pathVal == "" {
			return "工具 read_file 调用失败：缺少字符串参数 path", true
		}
		absPath, err := filepath.Abs(pathVal)
		if err != nil {
			return fmt.Sprintf("工具 read_file 调用失败：无法解析路径 %v", err), true
		}
		f, err := os.Open(absPath)
		if err != nil {
			return fmt.Sprintf("工具 read_file 调用失败：%v", err), true
		}
		defer f.Close()

		data, err := io.ReadAll(io.LimitReader(f, 4096))
		if err != nil {
			return fmt.Sprintf("工具 read_file 读取失败：%v", err), true
		}
		return fmt.Sprintf("文件 %s 的前 4KB 内容如下：\n%s", absPath, string(data)), true
	default:
		return fmt.Sprintf("未知工具：%s。可用工具：get_time, read_file。", tc.Tool), true
	}
}

// CallLLM sends the conversation to the model and returns its reply.
func (a *Agent) CallLLM(ctx context.Context) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: a.SystemPrompt,
		},
	}
	messages = append(messages, a.History...)

	reqBody := ChatCompletionRequest{
		Model:    a.Model,
		Messages: messages,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.APIURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("LLM API 返回错误状态码: %d", resp.StatusCode)
	}

	var cr ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("LLM 返回结果为空")
	}

	return cr.Choices[0].Message.Content, nil
}

func main() {
	// 从 .env 加载环境变量（若存在）
	if err := env.LoadEnv(); err != nil {
		fmt.Fprintln(os.Stderr, "加载 .env 失败:", err)
		os.Exit(1)
	}

	// 定义 Agent 的“人格”和默认行为。
	systemPrompt := `
你是一个用 Go 写的命令行 AI Agent Demo。
- 用简洁、直接的方式回答用户问题。
- 适当分点说明思路。
- 如果用户问到编程问题，尽量给出可运行的示例代码。
- 不确定的事情要说明不确定，而不是编造事实。

你可以在需要时调用以下“工具”，但要遵守协议：

可用工具：
- get_time: 获取当前本地时间。
  调用格式（必须是纯 JSON，单独一段，不要加多余文字）：
  {"tool":"get_time","args":{}}

- read_file: 读取一个本地文本文件的前 4KB 内容。
  调用格式（必须是纯 JSON，单独一段，不要加多余文字）：
  {"tool":"read_file","args":{"path":"./README.md"}}

当你决定使用工具时，请只输出上述 JSON，不要包含其它任何文字；否则就像普通聊天一样用中文正常回答。`

	agent := NewAgent(systemPrompt)
	ctx := context.Background()
	agent.Run(ctx)
}
