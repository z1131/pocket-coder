// Package api 封装与服务器的 HTTP API 交互（账号直连版）
package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "runtime"
    "time"
)

// Client API 客户端
// baseURL: 例如 http://localhost:8080
// accessToken: 部分接口需要鉴权时传入（Bearer）
type Client struct {
    baseURL    string
    httpClient *http.Client
}

// NewClient 创建 API 客户端
func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

// --- 通用响应 ---
type APIResponse struct {
    Code    int             `json:"code"`
    Message string          `json:"message"`
    Data    json.RawMessage `json:"data"`
}

// --- 认证 ---
type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

// Login 使用用户名密码登录
func (c *Client) Login(username, password string) (*LoginResponse, error) {
    body := map[string]string{
        "username": username,
        "password": password,
    }
    resp, err := c.post("/api/v1/auth/login", body, "")
    if err != nil {
        return nil, err
    }
    var result LoginResponse
    if err := json.Unmarshal(resp.Data, &result); err != nil {
        return nil, fmt.Errorf("解析登录响应失败: %w", err)
    }
    return &result, nil
}

// --- 设备注册 ---
type RegisterDesktopRequest struct {
    Name       string  `json:"name"`
    DeviceUUID string  `json:"device_uuid"`            // 设备唯一标识（持久化的 UUID）
    AgentType  *string `json:"agent_type,omitempty"`
    WorkingDir *string `json:"working_dir,omitempty"`
    OSInfo     *string `json:"os_info,omitempty"`
}

type RegisterDesktopResponse struct {
    DesktopID    int64   `json:"desktop_id"`
    DesktopToken string  `json:"desktop_token"`
    Name         string  `json:"name"`
    AgentType    string  `json:"agent_type"`
    OSInfo       *string `json:"os_info"`
    WorkingDir   *string `json:"working_dir"`
}

// RegisterDesktop 登录后注册桌面，返回 desktop_id 与桌面专用 token
func (c *Client) RegisterDesktop(accessToken string, req *RegisterDesktopRequest) (*RegisterDesktopResponse, error) {
    if req == nil {
        return nil, fmt.Errorf("请求体为空")
    }
    resp, err := c.post("/api/v1/desktops/register", req, accessToken)
    if err != nil {
        return nil, err
    }
    var result RegisterDesktopResponse
    if err := json.Unmarshal(resp.Data, &result); err != nil {
        return nil, fmt.Errorf("解析注册响应失败: %w", err)
    }
    return &result, nil
}

// --- 通用请求封装 ---
func (c *Client) get(path string, accessToken string) (*APIResponse, error) {
    url := c.baseURL + path
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    if accessToken != "" {
        req.Header.Set("Authorization", "Bearer "+accessToken)
    }
    return c.do(req)
}

func (c *Client) post(path string, body interface{}, accessToken string) (*APIResponse, error) {
    url := c.baseURL + path
    jsonBody, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    if accessToken != "" {
        req.Header.Set("Authorization", "Bearer "+accessToken)
    }
    return c.do(req)
}

func (c *Client) do(req *http.Request) (*APIResponse, error) {
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("读取响应失败: %w", err)
    }

    var apiResp APIResponse
    if err := json.Unmarshal(respBody, &apiResp); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    if apiResp.Code != 0 {
        return nil, fmt.Errorf("API 错误: %s", apiResp.Message)
    }

    return &apiResp, nil
}

// getHostname 获取主机名
func getHostname() string {
    hostname, _ := os.Hostname()
    if hostname == "" {
        hostname = "unknown"
    }
    return hostname
}

// getOSInfo 获取操作系统信息
func getOSInfo() string {
    return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
