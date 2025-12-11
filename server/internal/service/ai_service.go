// Package service 提供业务逻辑层的实现
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIService AI 服务
// 使用 Qwen API 进行会话总结
type AIService struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// NewAIService 创建 AIService 实例
func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey:     apiKey,
		apiURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// QwenRequest Qwen API 请求结构
type QwenRequest struct {
	Model    string         `json:"model"`
	Messages []QwenMessage  `json:"messages"`
}

// QwenMessage Qwen 消息结构
type QwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QwenResponse Qwen API 响应结构
type QwenResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SessionSummary 会话总结结果
type SessionSummary struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// GenerateSessionSummary 根据对话内容生成会话标题和摘要
// 参数:
//   - ctx: 上下文
//   - messages: 对话消息列表，格式为 [{role: "user/assistant", content: "..."}]
//
// 返回:
//   - *SessionSummary: 包含标题和摘要
//   - error: API 调用错误
func (s *AIService) GenerateSessionSummary(ctx context.Context, messages []map[string]string) (*SessionSummary, error) {
	if len(messages) == 0 {
		return &SessionSummary{
			Title:   "新会话",
			Summary: "暂无内容",
		}, nil
	}

	// 构建对话摘要请求
	var conversationText string
	for _, msg := range messages {
		role := msg["role"]
		content := msg["content"]
		if role == "user" {
			conversationText += fmt.Sprintf("用户: %s\n", content)
		} else {
			conversationText += fmt.Sprintf("AI: %s\n", content)
		}
	}

	// 限制对话文本长度，避免超出 token 限制
	if len(conversationText) > 4000 {
		conversationText = conversationText[:4000] + "..."
	}

	prompt := fmt.Sprintf(`请根据以下对话内容，生成一个简洁的标题和摘要。

对话内容:
%s

请按以下 JSON 格式返回（不要包含其他内容）:
{
  "title": "简洁的标题，不超过20字，概括对话主题",
  "summary": "简洁的摘要，不超过100字，说明对话的主要内容和目的"
}`, conversationText)

	// 调用 Qwen API
	reqBody := QwenRequest{
		Model: "qwen-turbo",
		Messages: []QwenMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var qwenResp QwenResponse
	if err := json.Unmarshal(body, &qwenResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if qwenResp.Error != nil {
		return nil, fmt.Errorf("API 错误: %s", qwenResp.Error.Message)
	}

	if len(qwenResp.Choices) == 0 {
		return nil, fmt.Errorf("API 返回空结果")
	}

	// 解析返回的 JSON
	content := qwenResp.Choices[0].Message.Content
	var summary SessionSummary
	if err := json.Unmarshal([]byte(content), &summary); err != nil {
		// 如果解析失败，使用默认值
		return &SessionSummary{
			Title:   "编程会话",
			Summary: content,
		}, nil
	}

	return &summary, nil
}
