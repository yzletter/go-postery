package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// RepoInfo 仓库信息
type RepoInfo struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Stars int    `json:"stars"`
	Desc  string `json:"desc"`
}

// GitHubSearcher 基于 GitHub MCP Server 的项目搜索器
type GitHubSearcher struct {
	client *client.Client
	mu     sync.Mutex
}

// NewGitHubSearcher 创建 GitHub 搜索器，启动 GitHub MCP Server（stdio 方式）
// token 为 GitHub Personal Access Token
func NewGitHubSearcher(token string) (*GitHubSearcher, error) {
	env := []string{fmt.Sprintf("GITHUB_PERSONAL_ACCESS_TOKEN=%s", token)}

	cli, err := client.NewStdioMCPClient("npx", env, "@modelcontextprotocol/server-github")
	if err != nil {
		return nil, fmt.Errorf("mcp/github: 启动 GitHub MCP Server 失败: %w", err)
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
		return nil, fmt.Errorf("mcp/github: MCP 初始化失败: %w", err)
	}

	log.Println("[MCP] GitHub Searcher 就绪")
	return &GitHubSearcher{client: cli}, nil
}

// SearchRepos 搜索 GitHub 开源项目
// query: GitHub 搜索语法，如 "go interview stars:>100"
// limit: 返回结果数量上限
func (gs *GitHubSearcher) SearchRepos(ctx context.Context, query string, limit int) ([]RepoInfo, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	req := mcp.CallToolRequest{}
	req.Params.Name = "search_repositories"
	req.Params.Arguments = map[string]any{
		"query":   query,
		"perPage": limit,
	}

	result, err := gs.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mcp/github: 搜索失败: %w", err)
	}

	// 解析返回的文本内容（GitHub MCP 返回 JSON 格式的搜索结果）
	var repos []RepoInfo
	for _, c := range result.Content {
		tc, ok := c.(mcp.TextContent)
		if !ok {
			continue
		}

		// 尝试解析为 GitHub API 搜索结果格式
		var searchResult struct {
			Items []struct {
				FullName        string `json:"full_name"`
				HTMLURL         string `json:"html_url"`
				Description     string `json:"description"`
				StargazersCount int    `json:"stargazers_count"`
			} `json:"items"`
		}

		if err := json.Unmarshal([]byte(tc.Text), &searchResult); err != nil {
			// 如果不是标准 JSON 格式，尝试直接解析为仓库数组
			var items []struct {
				FullName        string `json:"full_name"`
				HTMLURL         string `json:"html_url"`
				Description     string `json:"description"`
				StargazersCount int    `json:"stargazers_count"`
			}
			if err2 := json.Unmarshal([]byte(tc.Text), &items); err2 != nil {
				log.Printf("[MCP/GitHub] 解析搜索结果失败，原始内容: %s", tc.Text[:min(200, len(tc.Text))])
				continue
			}
			for _, item := range items {
				repos = append(repos, RepoInfo{
					Name:  item.FullName,
					URL:   item.HTMLURL,
					Stars: item.StargazersCount,
					Desc:  item.Description,
				})
			}
			continue
		}

		for _, item := range searchResult.Items {
			repos = append(repos, RepoInfo{
				Name:  item.FullName,
				URL:   item.HTMLURL,
				Stars: item.StargazersCount,
				Desc:  item.Description,
			})
		}
	}

	return repos, nil
}

// Close 关闭 MCP 客户端
func (gs *GitHubSearcher) Close() {
	// stdio 客户端关闭时会自动终止子进程
}
