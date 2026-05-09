package parser

import (
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    *URLInfo
		wantErr bool
	}{
		// ==================== 合法 Issue URL ====================
		{
			name:   "valid issue URL",
			rawURL: "https://github.com/owner/repo/issues/42",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 42,
			},
			wantErr: false,
		},
		{
			name:   "issue URL with trailing slash",
			rawURL: "https://github.com/owner/repo/issues/42/",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 42,
			},
			wantErr: false,
		},
		{
			name:   "issue URL with query parameters",
			rawURL: "https://github.com/owner/repo/issues/42?foo=bar&page=1",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 42,
			},
			wantErr: false,
		},
		{
			name:   "issue URL with fragment",
			rawURL: "https://github.com/owner/repo/issues/42#issuecomment-123",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 42,
			},
			wantErr: false,
		},
		{
			name:   "issue URL with large number",
			rawURL: "https://github.com/golang/go/issues/99999",
			want: &URLInfo{
				Owner:  "golang",
				Repo:   "go",
				Type:   ResourceIssue,
				Number: 99999,
			},
			wantErr: false,
		},

		// ==================== 合法 Pull Request URL ====================
		{
			name:   "valid pull request URL",
			rawURL: "https://github.com/owner/repo/pull/128",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourcePull,
				Number: 128,
			},
			wantErr: false,
		},
		{
			name:   "pull request URL with trailing slash",
			rawURL: "https://github.com/owner/repo/pull/128/",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourcePull,
				Number: 128,
			},
			wantErr: false,
		},
		{
			name:   "pull request with single-digit number",
			rawURL: "https://github.com/kubernetes/kubernetes/pull/1",
			want: &URLInfo{
				Owner:  "kubernetes",
				Repo:   "kubernetes",
				Type:   ResourcePull,
				Number: 1,
			},
			wantErr: false,
		},

		// ==================== 合法 Discussion URL ====================
		{
			name:   "valid discussion URL",
			rawURL: "https://github.com/owner/repo/discussions/99",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceDiscussion,
				Number: 99,
			},
			wantErr: false,
		},
		{
			name:   "discussion URL with trailing slash",
			rawURL: "https://github.com/owner/repo/discussions/99/",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceDiscussion,
				Number: 99,
			},
			wantErr: false,
		},
		{
			name:   "discussion with mixed case org name",
			rawURL: "https://github.com/MyOrg/my-repo/discussions/5",
			want: &URLInfo{
				Owner:  "MyOrg",
				Repo:   "my-repo",
				Type:   ResourceDiscussion,
				Number: 5,
			},
			wantErr: false,
		},

		// ==================== 无效 URL（格式错误） ====================
		{
			name:    "completely invalid URL",
			rawURL:  "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty string",
			rawURL:  "",
			wantErr: true,
		},
		{
			name:    "random path without host",
			rawURL:  "/owner/repo/issues/42",
			wantErr: true,
		},
		{
			name:    "non-http scheme",
			rawURL:  "ftp://github.com/owner/repo/issues/1",
			wantErr: true,
		},

		// ==================== 不支持的 URL 类型 ====================
		{
			name:    "repo home page URL",
			rawURL:  "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "repo tree URL",
			rawURL:  "https://github.com/owner/repo/tree/main",
			wantErr: true,
		},
		{
			name:    "repo blob URL",
			rawURL:  "https://github.com/owner/repo/blob/main/README.md",
			wantErr: true,
		},
		{
			name:    "repo wiki URL",
			rawURL:  "https://github.com/owner/repo/wiki",
			wantErr: true,
		},
		{
			name:    "non-GitHub host",
			rawURL:  "https://gitlab.com/owner/repo/issues/1",
			wantErr: true,
		},
		{
			name:    "GitHub subdomain",
			rawURL:  "https://mycompany.github.com/owner/repo/issues/1",
			wantErr: true,
		},

		// ==================== 边界情况 ====================
		{
			name:    "non-numeric issue number",
			rawURL:  "https://github.com/owner/repo/issues/abc",
			wantErr: true,
		},
		{
			name:   "number zero in URL",
			rawURL: "https://github.com/owner/repo/issues/0",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 0,
			},
			wantErr: false,
		},
		{
			name:    "negative number",
			rawURL:  "https://github.com/owner/repo/issues/-1",
			wantErr: true,
		},
		{
			name:    "unknown resource type",
			rawURL:  "https://github.com/owner/repo/releases/v1.0",
			wantErr: true,
		},
		{
			name:   "URL with www prefix",
			rawURL: "https://www.github.com/owner/repo/issues/42",
			want: &URLInfo{
				Owner:  "owner",
				Repo:   "repo",
				Type:   ResourceIssue,
				Number: 42,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.rawURL)

			// 验证错误期望
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseURL(%q) expected error, got %+v", tt.rawURL, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseURL(%q) unexpected error: %v", tt.rawURL, err)
			}

			// 验证字段值
			if got == nil {
				t.Fatal("ParseURL returned nil result without error")
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.Repo != tt.want.Repo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.want.Repo)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Number != tt.want.Number {
				t.Errorf("Number = %d, want %d", got.Number, tt.want.Number)
			}
		})
	}
}
