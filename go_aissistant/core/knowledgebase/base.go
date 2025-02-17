package knowledgebase

import "github.com/fighthorse/aicode/go_aissistant/config"

type KnowledgeBaseI interface {
	Initialize() error
	AddDocuments(docs []Document) error
	Query(query string, nResults int) ([]Document, error)
	DeleteDocument(id string) error
	ListDocuments() ([]Document, error)
}

type Document struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text"`
	Metadata map[string]interface{} `json:"metadata"`
}

type KnowledgeBaseManager struct {
	config    *config.AppConfig
	defaultKb KnowledgeBaseI
}

func NewKnowledgeBaseManager(conf *config.AppConfig) (*KnowledgeBaseManager, error) {
	manager := &KnowledgeBaseManager{
		config: conf,
	}

	// 根据配置初始化默认知识库

	chromaDb, err := NewChromaKB(conf)
	if err != nil {
		return nil, err
	}
	manager.defaultKb = chromaDb

	if err = manager.defaultKb.Initialize(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (km *KnowledgeBaseManager) Initialize() error {
	return km.defaultKb.Initialize()
}

// 添加文档到知识库
func (km *KnowledgeBaseManager) AddDocuments(docs []Document) error {
	return km.defaultKb.AddDocuments(docs)
}

// 查询知识库
func (km *KnowledgeBaseManager) Query(query string, numResults int) ([]Document, error) {
	return km.defaultKb.Query(query, numResults)
}

// 删除文档
func (km *KnowledgeBaseManager) DeleteDocument(id string) error {
	return km.defaultKb.DeleteDocument(id)
}

// 列出所有文档
func (km *KnowledgeBaseManager) ListDocuments() ([]Document, error) {
	return km.defaultKb.ListDocuments()
}
