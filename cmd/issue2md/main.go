package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bigwhite/issue2md/internal/cli"
	"github.com/bigwhite/issue2md/internal/config"
	"github.com/bigwhite/issue2md/internal/converter"
	"github.com/bigwhite/issue2md/internal/github"
	"github.com/bigwhite/issue2md/internal/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// 1. 解析命令行参数
	params, err := cli.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// 2. 解析 URL
	info, err := parser.ParseURL(params.URL)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	// 3. 创建 GitHub 客户端
	token := config.GetGitHubToken()
	client := github.NewClient(token)
	ctx := context.Background()

	// 4. 创建转换器
	conv := converter.New(converter.Option{
		EnableReactions: params.EnableReactions,
		EnableUserLinks: params.EnableUserLinks,
	})

	// 5. 根据资源类型获取并转换
	var md string
	switch info.Type {
	case parser.ResourceIssue:
		issue, err := client.FetchIssue(ctx, info.Owner, info.Repo, info.Number)
		if err != nil {
			return err
		}
		md, err = conv.ConvertIssue(issue)
		if err != nil {
			return err
		}
	case parser.ResourcePull:
		pr, err := client.FetchPullRequest(ctx, info.Owner, info.Repo, info.Number)
		if err != nil {
			return err
		}
		md, err = conv.ConvertPullRequest(pr)
		if err != nil {
			return err
		}
	case parser.ResourceDiscussion:
		disc, err := client.FetchDiscussion(ctx, info.Owner, info.Repo, info.Number)
		if err != nil {
			return err
		}
		md, err = conv.ConvertDiscussion(disc)
		if err != nil {
			return err
		}
	}

	// 6. 输出
	if params.OutputFile != "" {
		dir := filepath.Dir(params.OutputFile)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
		}
		if err := os.WriteFile(params.OutputFile, []byte(md), 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
	} else {
		fmt.Print(md)
	}

	return nil
}
