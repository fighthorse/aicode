package websearch

import (
	"context"
	"fmt"
	"github.com/fighthorse/aicode/go_aissistant/core/knowledgebase"
	"strings"
)

type WebSearchI interface {
	Search(ctx context.Context, query string, numResults int) ([]SearchResult, error)
}

type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

type BaseSearchClient struct {
	Name string
}

func (c *BaseSearchClient) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	return nil, nil
}

func NewBaseClient() *BaseSearchClient {
	fmt.Println("NewBaseClient")
	return &BaseSearchClient{
		Name: "BaseSearchClient",
	}
}

func BuildPrompt(query string, kbDocs []knowledgebase.Document, webResults []SearchResult) string {
	var builder strings.Builder

	// 知识库内容
	builder.WriteString("知识库参考内容：\n")
	for i, doc := range kbDocs {
		builder.WriteString(fmt.Sprintf("[知识%d] %s\n", i+1, doc.Text))
	}

	// 网络搜索结果
	builder.WriteString("\n网络搜索结果：\n")
	for i, result := range webResults {
		builder.WriteString(fmt.Sprintf("[网络%d] %s\n%s\n", i+1, result.Title, result.Snippet))
	}

	// 最终问题
	builder.WriteString(fmt.Sprintf("\n请根据以上信息回答：%s", query))

	return builder.String()
}
