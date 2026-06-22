package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// LLMConfig 从环境变量加载的 LLM 配置。
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// LoadLLMConfig 从环境变量读取 LLM 配置。
func LoadLLMConfig() LLMConfig {
	return LLMConfig{
		APIKey:  os.Getenv("LLM_API_KEY"),
		BaseURL: os.Getenv("LLM_BASE_URL"),
		Model:   os.Getenv("LLM_MODEL"),
	}
}

// ChatModel 统一的对话模型接口，兼容 OpenAI 和 Ollama。
type ChatModel struct {
	cfg        LLMConfig
	httpClient *http.Client
	isOllama   bool
}

// NewChatModel 创建 ChatModel 实例。
func NewChatModel(ctx context.Context, cfg LLMConfig) (*ChatModel, error) {
	return &ChatModel{
		cfg:        cfg,
		httpClient: &http.Client{},
		isOllama:   cfg.APIKey == "ollama",
	}, nil
}

// ---- OpenAI 兼容格式 ----

type openAIRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMsg     `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float32       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ---- Ollama 原生格式 ----

type ollamaRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMsg     `json:"messages"`
	Stream         bool          `json:"stream"`
	EnableThinking *bool         `json:"enable_thinking,omitempty"`
	Options        ollamaOptions `json:"options,omitempty"`
}

type ollamaOptions struct {
	Temperature float32 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

type ollamaResponse struct {
	Message chatMsg `json:"message"`
	Done    bool    `json:"done"`
}

// Generate 发送消息到 LLM 并返回回复。
func (cm *ChatModel) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	chatMsgs := make([]chatMsg, len(messages))
	for i, m := range messages {
		chatMsgs[i] = chatMsg{
			Role:    string(m.Role),
			Content: m.Content,
		}
	}

	if cm.isOllama {
		return cm.callOllama(ctx, chatMsgs)
	}
	return cm.callOpenAI(ctx, chatMsgs)
}

func (cm *ChatModel) callOllama(ctx context.Context, msgs []chatMsg) (*schema.Message, error) {
	f := false
	reqBody := ollamaRequest{
		Model:          cm.cfg.Model,
		Messages:       msgs,
		Stream:         false,
		EnableThinking: &f,
		Options: ollamaOptions{
			Temperature: 0.8,
			NumPredict:  512,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(cm.cfg.BaseURL, "/v1") + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") == "application/x-ndjson" {
		return cm.parseOllamaStream(resp.Body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama api error %d: %s", resp.StatusCode, string(respBody))
	}

	var or ollamaResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if or.Message.Content == "" {
		return nil, fmt.Errorf("empty response from ollama")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: or.Message.Content,
	}, nil
}

func (cm *ChatModel) callOpenAI(ctx context.Context, msgs []chatMsg) (*schema.Message, error) {
	reqBody := openAIRequest{
		Model:       cm.cfg.Model,
		Messages:    msgs,
		MaxTokens:   2048,
		Temperature: 0.8,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(cm.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cm.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cm.cfg.APIKey)
	}

	resp, err := cm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm api error %d: %s", resp.StatusCode, string(respBody))
	}

	var cr openAIResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if cr.Error != nil {
		return nil, fmt.Errorf("llm error: %s", cr.Error.Message)
	}

	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	content := cr.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("empty response content")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}, nil
}

// parseOllamaStream 解析 Ollama NDJSON 流式响应。
func (cm *ChatModel) parseOllamaStream(body io.Reader) (*schema.Message, error) {
	var fullContent strings.Builder
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var chunk ollamaResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	content := fullContent.String()
	if content == "" {
		return nil, fmt.Errorf("empty stream response")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}, nil
}
