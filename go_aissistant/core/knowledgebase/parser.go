package knowledgebase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type FileParser struct {
	SupportedFormats map[string]bool
}

func NewFileParser() *FileParser {
	return &FileParser{
		SupportedFormats: map[string]bool{
			".txt":  true,
			".pdf":  true,
			".docx": true,
		},
	}
}

func (p *FileParser) ParseFile(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".txt":
		return p.parseTextFile(path)
	case ".pdf":
		return p.parsePDFFile(path)
	case ".docx":
		return p.parseDocxFile(path)
	default:
		return "", fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

func (p *FileParser) parseTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (p *FileParser) parsePDFFile(path string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "pdfextract")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	if err := api.ExtractContentFile(path, tmpDir, nil, nil); err != nil {
		return "", err
	}

	var builder strings.Builder
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".txt") {
			content, _ := os.ReadFile(path)
			builder.Write(content)
			builder.WriteString("\n\n")
		}
		return nil
	})

	return builder.String(), nil
}

func (p *FileParser) parseDocxFile(path string) (string, error) {
	// 使用gooxml等库实现DOCX解析
	// 此处为示例占位实现
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
