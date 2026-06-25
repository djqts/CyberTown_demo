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
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func LoadLLMConfig() LLMConfig {
	cfg := LLMConfig{
		APIKey:  os.Getenv("LLM_API_KEY"),
		BaseURL: os.Getenv("LLM_BASE_URL"),
		Model:   os.Getenv("LLM_MODEL"),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5:3b"
	}
	if cfg.APIKey == "" && strings.Contains(cfg.BaseURL, "localhost:11434") {
		cfg.APIKey = "ollama"
	}
	return cfg
}

type ChatModel struct {
	cfg        LLMConfig
	httpClient *http.Client
	isOllama   bool
}

func NewChatModel(_ context.Context, cfg LLMConfig) (*ChatModel, error) {
	return &ChatModel{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		isOllama: cfg.APIKey == "ollama",
	}, nil
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []chatMsg `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream"`
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

func (cm *ChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...einomodel.Option) (*schema.Message, error) {
	commonOpts := einomodel.GetCommonOptions(&einomodel.Options{}, opts...)
	chatMsgs := make([]chatMsg, len(messages))
	for i, m := range messages {
		chatMsgs[i] = chatMsg{
			Role:    string(m.Role),
			Content: m.Content,
		}
	}

	if cm.isOllama {
		return cm.callOllama(ctx, chatMsgs, commonOpts)
	}
	return cm.callOpenAI(ctx, chatMsgs, commonOpts)
}

func (cm *ChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	resp, err := cm.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{resp}), nil
}

func (cm *ChatModel) callOllama(ctx context.Context, msgs []chatMsg, opts *einomodel.Options) (*schema.Message, error) {
	f := false
	reqBody := ollamaRequest{
		Model:          optionModel(opts, cm.cfg.Model),
		Messages:       msgs,
		Stream:         false,
		EnableThinking: &f,
		Options: ollamaOptions{
			Temperature: optionTemperature(opts, 0.8),
			NumPredict:  optionMaxTokens(opts, 512),
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(cm.cfg.BaseURL, "/v1") + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/x-ndjson") {
		return cm.parseOllamaStream(resp.Body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama api error %d: %s", resp.StatusCode, string(respBody))
	}

	var or ollamaResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if or.Message.Content == "" {
		return nil, fmt.Errorf("empty response from ollama")
	}
	return schema.AssistantMessage(or.Message.Content, nil), nil
}

func (cm *ChatModel) callOpenAI(ctx context.Context, msgs []chatMsg, opts *einomodel.Options) (*schema.Message, error) {
	reqBody := openAIRequest{
		Model:       optionModel(opts, cm.cfg.Model),
		Messages:    msgs,
		MaxTokens:   optionMaxTokens(opts, 256),
		Temperature: optionTemperature(opts, 0.7),
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(cm.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
	if resp.StatusCode != http.StatusOK {
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
	return schema.AssistantMessage(content, nil), nil
}

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
	return schema.AssistantMessage(content, nil), nil
}

func optionModel(opts *einomodel.Options, fallback string) string {
	if opts != nil && opts.Model != nil {
		return *opts.Model
	}
	return fallback
}

func optionTemperature(opts *einomodel.Options, fallback float32) float32 {
	if opts != nil && opts.Temperature != nil {
		return *opts.Temperature
	}
	return fallback
}

func optionMaxTokens(opts *einomodel.Options, fallback int) int {
	if opts != nil && opts.MaxTokens != nil {
		return *opts.MaxTokens
	}
	return fallback
}
