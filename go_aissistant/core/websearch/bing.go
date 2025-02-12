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

type BingSearchClient struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewBingSearchClient(apiKey string) *BingSearchClient {
	return &BingSearchClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *BingSearchClient) Search(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	baseURL := "https://api.bing.microsoft.com/v7.0/search"

	params := url.Values{}
	params.Add("q", query)
	params.Add("count", fmt.Sprintf("%d", numResults))

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		log.Printf("创建请求时出错: %v", err)
		return nil, err
	}

	req.Header.Add("Ocp-Apim-Subscription-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("发送请求时出错: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bing API 返回错误: %s", resp.Status)
		return nil, fmt.Errorf("bing API 返回错误: %s", resp.Status)
	}

	var result struct {
		WebPages struct {
			Value []struct {
				Name    string `json:"name"`
				URL     string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
		} `json:"webPages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("解码响应时出错: %v", err)
		return nil, err
	}

	var results []SearchResult
	for _, item := range result.WebPages.Value {
		results = append(results, SearchResult{
			Title:   item.Name,
			Link:    item.URL,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
