package config

import "os"

// EnvGitHubToken 是 GitHub Personal Access Token 的环境变量名。
const EnvGitHubToken = "GITHUB_TOKEN"

// GetGitHubToken 从环境变量读取 GitHub Token。
// 未设置环境变量时返回空字符串。
func GetGitHubToken() string {
	return os.Getenv(EnvGitHubToken)
}
