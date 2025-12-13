package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"pocket-coder-server/internal/config"
)

const (
	// DashScope API Endpoint
	QwenEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	// Model Name
	QwenModel = "qwen-turbo"
)

// AIService 提供 AI 相关功能
type AIService struct {
	config *config.Config
	client *http.Client
}

// NewAIService 创建 AIService 实例
func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second, // 设置超时
		},
	}
}

// GenerateCommandRequest 命令生成请求
type GenerateCommandRequest struct {
	Prompt  string `json:"prompt"`
	Context struct {
		OS    string `json:"os"`
		Shell string `json:"shell"`
	} `json:"context"`
}

// GenerateCommandResponse 命令生成响应
type GenerateCommandResponse struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}

// DashScopeRequest 阿里云 API 请求结构
type DashScopeRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []DashScopeMessage `json:"messages"`
	} `json:"input"`
	Parameters struct {
		ResultFormat string `json:"result_format"` // "message"
	} `json:"parameters"`
}

type DashScopeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DashScopeResponse 阿里云 API 响应结构
type DashScopeResponse struct {
	Output struct {
		Choices []struct {
			Message DashScopeMessage `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GenerateCommand 调用 Qwen 生成 Shell 命令
func (s *AIService) GenerateCommand(ctx context.Context, req *GenerateCommandRequest) (*GenerateCommandResponse, error) {
	if s.config.AI.QwenAPIKey == "" {
		return nil, errors.New("AI service not configured (missing API Key)")
	}

	// 1. 构建 System Prompt
	systemPrompt := "You are a strict shell command generator assistant.\n" +
		"Your goal is to translate natural language requests into precise shell commands.\n" +
		"Rules:\n" +
		"1. Output ONLY the shell command. Do not use markdown code blocks (```).\n" +
		"2. If an explanation is absolutely necessary or requested, put it after the command, separated by ' # '.\n" +
		"3. Be concise and safe.\n"

	if req.Context.OS != "" {
		systemPrompt += fmt.Sprintf("Target OS: %s.\n", req.Context.OS)
	}
	if req.Context.Shell != "" {
		systemPrompt += fmt.Sprintf("Target Shell: %s.\n", req.Context.Shell)
	}

	// 2. 构造请求 Body
	dashReq := DashScopeRequest{
		Model: QwenModel,
	}
	dashReq.Input.Messages = []DashScopeMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Prompt},
	}
	dashReq.Parameters.ResultFormat = "message"

	jsonData, err := json.Marshal(dashReq)
	if err != nil {
		return nil, err
	}

	// 3. 发送 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", QwenEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.AI.QwenAPIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 4. 解析响应
	var dashResp DashScopeResponse
	if err := json.Unmarshal(bodyBytes, &dashResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if dashResp.Code != "" {
		return nil, fmt.Errorf("AI service error: %s - %s", dashResp.Code, dashResp.Message)
	}

	if len(dashResp.Output.Choices) == 0 {
		return nil, errors.New("AI returned no content")
	}

	rawContent := dashResp.Output.Choices[0].Message.Content
	rawContent = strings.TrimSpace(rawContent)

	// 5. 后处理 (移除可能存在的 Markdown 标记，尽管 Prompt 要求不要有)
	rawContent = strings.TrimPrefix(rawContent, "```bash")
	rawContent = strings.TrimPrefix(rawContent, "```sh")
	rawContent = strings.TrimPrefix(rawContent, "```")
	rawContent = strings.TrimSuffix(rawContent, "```")
	rawContent = strings.TrimSpace(rawContent)

	// 分离命令和解释
	parts := strings.SplitN(rawContent, " # ", 2)
	result := &GenerateCommandResponse{
		Command: parts[0],
	}
	if len(parts) > 1 {
		result.Explanation = parts[1]
	}

	return result, nil
}
