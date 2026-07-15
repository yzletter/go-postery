// Package mcp 实现 MCP（Model Context Protocol）工具集成
// 通过 Playwright MCP Server 实现 JS 渲染网页抓取
package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// WebScraper 基于 Playwright MCP Server 的网页抓取器
type WebScraper struct {
	client *client.Client
	mu     sync.Mutex
}

// NewWebScraper 创建网页抓取器，启动 Playwright MCP Server（stdio 方式）
func NewWebScraper() (*WebScraper, error) {
	cli, err := client.NewStdioMCPClient(
		"npx", nil,
		"@playwright/mcp@latest", "--headless",
	)
	if err != nil {
		return nil, fmt.Errorf("mcp/web_scraper: 启动 Playwright MCP Server 失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "interview-agent",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initReq)
	if err != nil {
		return nil, fmt.Errorf("mcp/web_scraper: MCP 初始化失败: %w", err)
	}

	log.Println("[MCP] Playwright Web Scraper 就绪")
	return &WebScraper{client: cli}, nil
}

// ScrapeURL 通过 Playwright MCP 抓取网页内容（支持 JS 渲染）
// 返回页面的 accessibility snapshot 文本
func (ws *WebScraper) ScrapeURL(ctx context.Context, url string) (string, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// 1. 导航到目标 URL
	navReq := mcp.CallToolRequest{}
	navReq.Params.Name = "browser_navigate"
	navReq.Params.Arguments = map[string]any{
		"url": url,
	}

	_, err := ws.client.CallTool(ctx, navReq)
	if err != nil {
		return "", fmt.Errorf("mcp/web_scraper: 导航失败 (%s): %w", url, err)
	}

	// 2. 等待页面加载完成
	waitReq := mcp.CallToolRequest{}
	waitReq.Params.Name = "browser_wait_for"
	waitReq.Params.Arguments = map[string]any{
		"time": 3000, // 等待 3 秒让 JS 渲染完成
	}
	_, _ = ws.client.CallTool(ctx, waitReq)

	// 3. 获取页面 snapshot（accessibility tree，比 screenshot 更适合提取文本）
	snapReq := mcp.CallToolRequest{}
	snapReq.Params.Name = "browser_snapshot"
	snapReq.Params.Arguments = map[string]any{}

	snapResult, err := ws.client.CallTool(ctx, snapReq)
	if err != nil {
		return "", fmt.Errorf("mcp/web_scraper: 页面快照失败: %w", err)
	}

	// 4. 提取文本内容
	var content string
	for _, c := range snapResult.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			content += tc.Text + "\n"
		}
	}

	content = strings.TrimSpace(content)
	if len(content) < 50 {
		return "", fmt.Errorf("mcp/web_scraper: 页面内容过短，可能被反爬拦截（长度: %d）", len(content))
	}

	// 5. 检测浏览器错误页（连接失败、超时、DNS 解析失败等）
	lower := strings.ToLower(content)
	browserErrors := []string{
		"err_connection_closed", "err_connection_refused", "err_connection_reset",
		"err_connection_timed_out", "err_name_not_resolved", "err_internet_disconnected",
		"err_ssl_protocol_error", "err_cert_", "net::err_",
		"this site can't be reached", "无法访问此网站",
	}
	for _, errStr := range browserErrors {
		if strings.Contains(lower, errStr) {
			return "", fmt.Errorf("mcp/web_scraper: 网页无法访问（%s），请检查链接或改用粘贴文本方式", errStr)
		}
	}

	// 6. 检测是否被安全验证拦截
	if strings.Contains(lower, "security-check") ||
		strings.Contains(lower, "请稍候") ||
		strings.Contains(lower, "验证码") ||
		strings.Contains(lower, "正在加载中") {
		return "", fmt.Errorf("mcp/web_scraper: 页面触发了安全验证，无法自动抓取，请改用文件或手动输入方式")
	}

	return content, nil
}

// Close 关闭 MCP 客户端
func (ws *WebScraper) Close() {
	if ws.client != nil {
		// 关闭浏览器
		closeReq := mcp.CallToolRequest{}
		closeReq.Params.Name = "browser_close"
		closeReq.Params.Arguments = map[string]any{}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = ws.client.CallTool(ctx, closeReq)
	}
}
