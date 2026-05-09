package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ResourceType 表示 GitHub 资源类型。
type ResourceType string

const (
	ResourceIssue      ResourceType = "issue"
	ResourcePull       ResourceType = "pull"
	ResourceDiscussion ResourceType = "discussion"
)

// URLInfo 存储从 GitHub URL 解析出的结构化信息。
type URLInfo struct {
	Owner  string
	Repo   string
	Type   ResourceType
	Number int
}

// ParseURL 解析 GitHub URL，自动识别资源类型。
//
// 支持的格式:
//
//	https://github.com/{owner}/{repo}/issues/{number}
//	https://github.com/{owner}/{repo}/pull/{number}
//	https://github.com/{owner}/{repo}/discussions/{number}
//
// URL 允许尾部斜杠和无关查询参数。
func ParseURL(rawURL string) (*URLInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidURL, err)
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, ErrInvalidURL
	}

	host := u.Host
	if host == "www.github.com" {
		host = "github.com"
	}
	if host != "github.com" {
		return nil, ErrInvalidURL
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 {
		return nil, ErrInvalidURL
	}

	owner, repo := parts[0], parts[1]
	typeStr := parts[2]
	numStr := parts[3]

	var rtype ResourceType
	switch typeStr {
	case "issues":
		rtype = ResourceIssue
	case "pull":
		rtype = ResourcePull
	case "discussions":
		rtype = ResourceDiscussion
	default:
		return nil, ErrInvalidURL
	}

	number, err := strconv.Atoi(numStr)
	if err != nil || number < 0 {
		return nil, fmt.Errorf("%w: invalid issue/PR number", ErrInvalidURL)
	}

	return &URLInfo{
		Owner:  owner,
		Repo:   repo,
		Type:   rtype,
		Number: number,
	}, nil
}
