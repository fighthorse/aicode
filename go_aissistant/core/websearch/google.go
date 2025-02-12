package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type GoogleSearchClient struct {
	APIKey     string
	CX         string
	HTTPClient *http.Client
}

func NewGoogleSearchClient(apiKey, cx string) *GoogleSearchClient {
	return &GoogleSearchClient{
		APIKey: apiKey,
		CX:     cx,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *GoogleSearchClient) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"

	params := url.Values{}
	params.Add("key", c.APIKey)
	params.Add("cx", c.CX)
	params.Add("q", query)
	params.Add("num", fmt.Sprintf("%d", numResults))

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		log.Printf("创建请求时出错: %v", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("发送请求时出错: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Google API 返回错误: %s", resp.Status)
		return nil, fmt.Errorf("google API 返回错误: %s", resp.Status)
	}

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("解码响应时出错: %v", err)
		return nil, err
	}

	var results []SearchResult
	for _, item := range result.Items {
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
