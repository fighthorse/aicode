package knowledgebase

import (
	"context"
	"fmt"
	"github.com/fighthorse/aicode/go_aissistant/config"
	"log"
	"sync"

	chroma "github.com/amikos-tech/chroma-go"
	openai "github.com/amikos-tech/chroma-go/pkg/embeddings/openai"
	"github.com/amikos-tech/chroma-go/types"
)

type Document struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}

type KnowledgeBase struct {
	client        *chroma.Client
	collection    *chroma.Collection
	collectionMu  sync.RWMutex
	embeddingFunc types.EmbeddingFunction
}

// NewKnowledgeBase creates a new KnowledgeBase instance
func NewKnowledgeBase(config *config.AppConfig) (*KnowledgeBase, error) {
	if config.ChromaURL == "" {
		config.ChromaURL = "http://localhost:8000"
	}

	// Create a new Chroma client
	client, err := chroma.NewClient(chroma.WithBasePath(config.ChromaURL))
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	embeddingFunc, err := openai.NewOpenAIEmbeddingFunction("text-embedding-3-small")
	if err != nil {
		log.Printf("创建嵌入函数时出错: %v", err)
		return nil, fmt.Errorf("创建嵌入函数时出错: %w", err)
	}

	return &KnowledgeBase{
		client:        client,
		embeddingFunc: embeddingFunc,
	}, nil
}

// Initialize initializes the knowledge base with a specific collection
func (kb *KnowledgeBase) Initialize(collectionName string) error {
	kb.collectionMu.Lock()
	defer kb.collectionMu.Unlock()

	ctx := context.Background()
	collections, err := kb.client.ListCollections(ctx)
	if err != nil {
		log.Printf("列出集合时出错: %v", err)
		return fmt.Errorf("列出集合时出错: %w", err)
	}

	for _, c := range collections {
		if c.Name == collectionName {
			kb.collection = c // 修复：去掉 &
			return nil
		}
	}

	// 创建新集合
	collection, err := kb.client.CreateCollection(ctx, collectionName, nil, true, kb.embeddingFunc, types.L2) // 假设正确的常量名称为 DistanceMetricCosine
	if err != nil {
		log.Printf("创建集合时出错: %v", err)
		return fmt.Errorf("创建集合时出错: %w", err)
	}
	kb.collection = collection
	return nil
}

// AddDocuments adds documents to the knowledge base
// AddDocuments adds documents to the knowledge base
func (kb *KnowledgeBase) AddDocuments(docs []Document) error {
	kb.collectionMu.Lock()
	defer kb.collectionMu.Unlock()

	ctx := context.Background()

	ids := make([]string, len(docs))
	documents := make([]string, len(docs))
	metadatas := make([]map[string]interface{}, len(docs))
	embeddings := make([]*types.Embedding, len(docs))

	for i, doc := range docs {
		ids[i] = doc.ID
		documents[i] = doc.Text
		metadatas[i] = doc.Metadata

		// 生成嵌入向量
		//embedding := kb.embeddingFunc(doc.Text)
		//embeddings[i] = embedding
	}

	// 调用 Add 方法时传递正确的参数类型
	_, err := kb.collection.Add(ctx, embeddings, metadatas, documents, ids) // 假设 Add 方法签名正确
	if err != nil {
		log.Printf("添加文档时出错: %v", err)
		return fmt.Errorf("添加文档时出错: %w", err)
	}
	return nil
}

// Query queries the knowledge base for documents
func (kb *KnowledgeBase) Query(query string, nResults int) ([]Document, error) {
	kb.collectionMu.RLock()
	defer kb.collectionMu.RUnlock()
	if kb.collection == nil {
		log.Printf("集合未初始化")
		return nil, fmt.Errorf("集合未初始化")
	}
	ctx := context.Background()
	results, err := kb.collection.Query(ctx, []string{query}, int32(nResults), nil, nil, nil)
	if err != nil {
		log.Printf("查询文档时出错: %v", err)
		return nil, fmt.Errorf("查询文档时出错: %w", err)
	}

	var docs []Document
	for i := range results.Ids {
		docs = append(docs, Document{
			ID:       string(results.Ids[i][0]),
			Text:     string(results.Documents[i][0]),
			Metadata: results.Metadatas[i][0],
		})
	}
	return docs, nil
}

// DeleteDocument deletes a document from the knowledge base
func (kb *KnowledgeBase) DeleteDocument(id string) error {
	kb.collectionMu.Lock()
	defer kb.collectionMu.Unlock()

	ctx := context.Background()
	deletedIds, err := kb.collection.Delete(ctx, []string{id}, nil, nil)
	if err != nil {
		log.Printf("删除文档时出错: %v", err)
		return fmt.Errorf("删除文档时出错: %w", err)
	}

	if len(deletedIds) == 0 {
		log.Printf("未找到要删除的文档: %s", id)
		return fmt.Errorf("未找到要删除的文档: %s", id)
	}
	return nil
}

// ListDocuments lists all documents in the knowledge base
func (kb *KnowledgeBase) ListDocuments() ([]Document, error) {
	kb.collectionMu.RLock()
	defer kb.collectionMu.RUnlock()

	ctx := context.Background()
	results, err := kb.collection.Get(ctx, nil, nil, nil, nil) // 假设 Get 方法用于列出文档，具体方法名需根据库文档确认
	if err != nil {
		log.Printf("列出文档时出错: %v", err)
		return nil, fmt.Errorf("列出文档时出错: %w", err)
	}

	var docs []Document
	for i := range results.Ids {
		docs = append(docs, Document{
			ID:       string(results.Ids[i][0]),       // 修复：将 byte 转换为 string
			Text:     string(results.Documents[i][0]), // 修复：将 byte 转换为 string
			Metadata: results.Metadatas[i],            // 修复：类型断言
		})
	}
	return docs, nil
}
