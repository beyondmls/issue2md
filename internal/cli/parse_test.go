package cli

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Params
		wantErr bool
	}{
		{
			name: "URL only",
			args: []string{"https://github.com/owner/repo/issues/1"},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "",
				EnableReactions: false,
				EnableUserLinks: false,
			},
			wantErr: false,
		},
		{
			name: "URL with output file",
			args: []string{"https://github.com/owner/repo/issues/1", "/tmp/issue.md"},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "/tmp/issue.md",
				EnableReactions: false,
				EnableUserLinks: false,
			},
			wantErr: false,
		},
		{
			name: "enable reactions flag",
			args: []string{"-enable-reactions", "https://github.com/owner/repo/issues/1"},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "",
				EnableReactions: true,
				EnableUserLinks: false,
			},
			wantErr: false,
		},
		{
			name: "enable user links flag",
			args: []string{"-enable-user-links", "https://github.com/owner/repo/issues/1"},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "",
				EnableReactions: false,
				EnableUserLinks: true,
			},
			wantErr: false,
		},
		{
			name: "both flags with output file",
			args: []string{
				"-enable-reactions",
				"-enable-user-links",
				"https://github.com/owner/repo/issues/1",
				"./output.md",
			},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "./output.md",
				EnableReactions: true,
				EnableUserLinks: true,
			},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "only flags no URL",
			args:    []string{"-enable-reactions"},
			wantErr: true,
		},
		{
			// Go 标准库 flag 在遇到第一个非 flag 参数后停止解析，
			// 因此 -enable-reactions 被视为输出文件路径参数。
			name: "flags after URL treated as positional args",
			args: []string{"https://github.com/owner/repo/issues/1", "-enable-reactions"},
			want: &Params{
				URL:             "https://github.com/owner/repo/issues/1",
				OutputFile:      "-enable-reactions",
				EnableReactions: false,
				EnableUserLinks: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.OutputFile != tt.want.OutputFile {
				t.Errorf("OutputFile = %q, want %q", got.OutputFile, tt.want.OutputFile)
			}
			if got.EnableReactions != tt.want.EnableReactions {
				t.Errorf("EnableReactions = %v, want %v", got.EnableReactions, tt.want.EnableReactions)
			}
			if got.EnableUserLinks != tt.want.EnableUserLinks {
				t.Errorf("EnableUserLinks = %v, want %v", got.EnableUserLinks, tt.want.EnableUserLinks)
			}
		})
	}
}
