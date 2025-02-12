package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type AppConfig struct {
	OllamaURL          string `json:"ollama_url"`
	GoogleAPIKey       string `json:"google_api_key"`
	GoogleCX           string `json:"google_cx"`
	BingAPIKey         string `json:"bing_api_key"`
	ChromaPath         string `json:"chroma_path"`
	ChromaURL          string `json:"chroma_url"`
	SQLitePath         string `json:"sqlite_path"` // 新增SQLite路径
	DefaultModel       string `json:"default_model"`
	HistoryLimit       int    `json:"history_limit"` // 历史记录条数限制
	RetentionDays      int    `json:"retention_days"`
	CollectionName     string `json:"collection_name"`
	EmbeddingModel     string `json:"embedding_model"`
	EmbeddingDimension int    `json:"embedding_dimension"`
	EmbeddingBatchSize int    `json:"embedding_batch_size"`
}

// LoadConfig 解析传入文件名称 ，通过json文件解析 返回appConfig配置及错误
func LoadConfig(filePath string) (*AppConfig, error) {
	// 路径验证
	if filePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}

	// 尝试打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	// 读取文件内容
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析JSON配置
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	log.Printf("Config loaded successfully from %s", filePath)
	return &config, nil
}

func SaveConfig(m *AppConfig) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
		return err
	}

	err = ioutil.WriteFile("config.json", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write config file: %v", err)
		return err
	}
	return nil
}
