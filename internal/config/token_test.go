package config

import (
	"os"
	"testing"
)

func TestGetGitHubToken_NotSet(t *testing.T) {
	// 确保环境变量不存在
	os.Unsetenv(EnvGitHubToken)
	token := GetGitHubToken()
	if token != "" {
		t.Errorf("expected empty string, got %q", token)
	}
}

func TestGetGitHubToken_Set(t *testing.T) {
	os.Setenv(EnvGitHubToken, "ghp_test123token")
	defer os.Unsetenv(EnvGitHubToken)

	token := GetGitHubToken()
	if token != "ghp_test123token" {
		t.Errorf("expected %q, got %q", "ghp_test123token", token)
	}
}

func TestGetGitHubToken_EmptyString(t *testing.T) {
	os.Setenv(EnvGitHubToken, "")
	token := GetGitHubToken()
	if token != "" {
		t.Errorf("expected empty string, got %q", token)
	}
}

func TestEnvConst(t *testing.T) {
	if EnvGitHubToken != "GITHUB_TOKEN" {
		t.Errorf("EnvGitHubToken = %q, want %q", EnvGitHubToken, "GITHUB_TOKEN")
	}
}
