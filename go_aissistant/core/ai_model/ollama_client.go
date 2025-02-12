package ai_model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type GenerationRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options"`
}

type Options struct {
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
}

type GenerationResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
	Created  string `json:"created_at"`
}

func (c *OllamaClient) Generate(prompt string, model string) (string, error) {
	reqBody := GenerationRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Options: Options{Temperature: 0.7, TopP: 0.9},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误: %s (%d)", string(body), resp.StatusCode)
	}

	var result GenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	return result.Response, nil
}

func (c *OllamaClient) ListLocalModels() ([]string, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range response.Models {
		models = append(models, m.Name)
	}
	return models, nil
}

func (c *OllamaClient) GenerateStream(prompt, model string, ch chan<- string) error {
	reqBody := GenerationRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  true,
		Options: Options{Temperature: 0.7},
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	go func() {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)

		for {
			var streamResp struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				close(ch)
				return
			}

			if streamResp.Done {
				close(ch)
				return
			}

			ch <- streamResp.Response
		}
	}()

	return nil
}
