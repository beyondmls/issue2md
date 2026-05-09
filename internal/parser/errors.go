package parser

import "errors"

// ErrInvalidURL 表示无法解析为有效的 GitHub Issue/PR/Discussion URL。
var ErrInvalidURL = errors.New("invalid GitHub URL")
