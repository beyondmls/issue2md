package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const graphqlEndpoint = "https://api.github.com/graphql"

// discussionQuery 是获取 Discussion 详细信息的 GraphQL Query。
// 一次性获取主楼、labels、comments 及 replies，保留层级结构。
const discussionQuery = `
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      id
      title
      url
      state
      body
      createdAt
      updatedAt
      author { login }
      category { name }
      labels(first: 10) { nodes { name } }
      comments(first: 100) {
        nodes {
          id
          body
          createdAt
          updatedAt
          author { login }
          isAnswer
          reactions(first: 10) { nodes { content } }
          replies(first: 100) {
            nodes {
              id
              body
              createdAt
              updatedAt
              author { login }
              reactions(first: 10) { nodes { content } }
            }
          }
        }
      }
    }
  }
}
`

// graphQLResponse 是 GraphQL API 的标准响应包装。
type graphQLResponse struct {
	Data   map[string]any `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// fetchDiscussionGraphQL 通过 GraphQL API 获取 Discussion 详情和评论。
func fetchDiscussionGraphQL(ctx context.Context, token, owner, repo string, number int) (*Discussion, error) {
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required to fetch discussions via GraphQL API")
	}

	variables := map[string]any{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	body, err := doGraphQL(ctx, token, discussionQuery, variables)
	if err != nil {
		return nil, fmt.Errorf("fetching discussion %s/%s#%d: %w", owner, repo, number, err)
	}

	disc, err := parseDiscussionResponse(body, number)
	if err != nil {
		return nil, fmt.Errorf("parsing discussion %s/%s#%d response: %w", owner, repo, number, err)
	}
	return disc, nil
}

// doGraphQL 发送 GraphQL 请求并返回原始 JSON body。
func doGraphQL(ctx context.Context, token, query string, variables map[string]any) ([]byte, error) {
	reqBody := map[string]any{
		"query":     query,
		"variables": variables,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading graphql response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("unmarshaling graphql response: %w", err)
	}
	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %s", gqlResp.Errors[0].Message)
	}

	return respBody, nil
}

// parseDiscussionResponse 从 GraphQL 响应中解析 Discussion 模型。
func parseDiscussionResponse(body []byte, number int) (*Discussion, error) {
	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	repoData, _ := gqlResp.Data["repository"].(map[string]any)
	if repoData == nil {
		return nil, fmt.Errorf("repository not found in response")
	}

	discData, _ := repoData["discussion"].(map[string]any)
	if discData == nil {
		return nil, fmt.Errorf("discussion not found in response")
	}

	disc := &Discussion{
		GitResource: GitHubResource{
			Number:    number,
			Title:     readString(discData, "title"),
			URL:       readString(discData, "url"),
			State:     toLower(readString(discData, "state")),
			Body:      readString(discData, "body"),
			CreatedAt: parseTime(readString(discData, "createdAt")),
			UpdatedAt: parseTime(readString(discData, "updatedAt")),
		},
	}

	// Author
	if author, ok := discData["author"].(map[string]any); ok {
		disc.GitResource.Author = readString(author, "login")
	}

	// Labels
	if labelsData, ok := discData["labels"].(map[string]any); ok {
		if nodes, ok := labelsData["nodes"].([]any); ok {
			for _, n := range nodes {
				if label, ok := n.(map[string]any); ok {
					disc.GitResource.Labels = append(disc.GitResource.Labels, readString(label, "name"))
				}
			}
		}
	}

	// Comments
	commentsData, _ := discData["comments"].(map[string]any)
	if commentsData != nil {
		nodes, _ := commentsData["nodes"].([]any)
		for _, n := range nodes {
			commentNode, ok := n.(map[string]any)
			if !ok {
				continue
			}
			parent := parseCommentNode(commentNode, nil)
			disc.Comments = append(disc.Comments, parent)

			// Replies (nested under comment's replies field)
			if repliesData, ok := commentNode["replies"].(map[string]any); ok {
				replyNodes, _ := repliesData["nodes"].([]any)
				parentID := parent.ID
				for _, rn := range replyNodes {
					replyNode, ok := rn.(map[string]any)
					if !ok {
						continue
					}
					reply := parseCommentNode(replyNode, &parentID)
					disc.Comments = append(disc.Comments, reply)
				}
			}
		}
	}

	// Sort comments by CreatedAt
	sortCommentsByCreatedAt(disc.Comments)

	return disc, nil
}

// parseCommentNode 从 GraphQL comment node map 中解析 Comment。
func parseCommentNode(node map[string]any, parentID *int64) Comment {
	c := Comment{
		Body:      readString(node, "body"),
		CreatedAt: parseTime(readString(node, "createdAt")),
		UpdatedAt: parseTime(readString(node, "updatedAt")),
		IsAnswer:  readBool(node, "isAnswer"),
		ParentID:  parentID,
	}

	if author, ok := node["author"].(map[string]any); ok {
		c.Author = readString(author, "login")
	}

	// Reactions
	if reactionsData, ok := node["reactions"].(map[string]any); ok {
		if nodes, ok := reactionsData["nodes"].([]any); ok {
			for _, rn := range nodes {
				rNode, ok := rn.(map[string]any)
				if !ok {
					continue
				}
				content := readString(rNode, "content")
				switch content {
				case "THUMBS_UP":
					c.Reactions.ThumbsUp++
				case "THUMBS_DOWN":
					c.Reactions.ThumbsDown++
				case "LAUGH":
					c.Reactions.Laugh++
				case "HOORAY":
					c.Reactions.Hooray++
				case "CONFUSED":
					c.Reactions.Confused++
				case "HEART":
					c.Reactions.Heart++
				case "ROCKET":
					c.Reactions.Rocket++
				case "EYES":
					c.Reactions.Eyes++
				}
			}
		}
	}

	return c
}

// --- helper functions ---

func readString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func readBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func toLower(s string) string {
	if s == "" {
		return ""
	}
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
