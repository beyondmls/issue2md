package cli

import (
	"errors"
	"flag"
)

// Params 存储命令行解析结果。
type Params struct {
	URL             string
	OutputFile      string // 空字符串表示输出到 stdout
	EnableReactions bool
	EnableUserLinks bool
}

// Parse 解析命令行参数。
//
// 用法: issue2md [flags] <url> [output_file]
//
// Flags:
//
//	-enable-reactions    包含 Reactions 统计
//	-enable-user-links   将 @user 渲染为 GitHub 链接
func Parse(args []string) (*Params, error) {
	fs := flag.NewFlagSet("issue2md", flag.ContinueOnError)

	enableReactions := fs.Bool("enable-reactions", false, "include reaction counts in output")
	enableUserLinks := fs.Bool("enable-user-links", false, "render @mentions as GitHub profile links")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if fs.NArg() < 1 {
		return nil, errors.New("usage: issue2md [flags] <url> [output_file]")
	}

	url := fs.Arg(0)
	outputFile := ""
	if fs.NArg() > 1 {
		outputFile = fs.Arg(1)
	}

	return &Params{
		URL:             url,
		OutputFile:      outputFile,
		EnableReactions: *enableReactions,
		EnableUserLinks: *enableUserLinks,
	}, nil
}
