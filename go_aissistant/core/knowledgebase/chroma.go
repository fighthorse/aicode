package knowledgebase

import (
	"context"
	"fmt"
	"github.com/amikos-tech/chroma-go/pkg/embeddings/ollama"
	"github.com/fighthorse/aicode/go_aissistant/config"
	"log"
	"sync"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"
)

type ChromaKB struct {
	collectionName string
	client         *chroma.Client
	collection     *chroma.Collection
	collectionMu   sync.RWMutex
	embeddingFunc  types.EmbeddingFunction
	metadata       map[string]interface{}
}

// NewKnowledgeBase creates a new KnowledgeBase instance
func NewChromaKB(config *config.AppConfig) (*ChromaKB, error) {
	if config.ChromaURL == "" {
		config.ChromaURL = "http://localhost:8000"
	}

	client, err := chroma.NewClient(chroma.WithBasePath(config.ChromaURL))
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	embeddingFunc, err := ollama.NewOllamaEmbeddingFunction(ollama.WithBaseURL("http://127.0.0.1:11434"), ollama.WithModel("nomic-embed-text"))
	if err != nil {
		fmt.Printf("Error creating Ollama embedding function: %s \n", err)
	}
	metadata := make(map[string]interface{})
	fmt.Println("NewChromaKB ==>", config.CollectionName)
	return &ChromaKB{
		collectionName: config.CollectionName,
		client:         client,
		embeddingFunc:  embeddingFunc,
		metadata:       metadata,
	}, nil
}

// Initialize initializes the knowledge base with a specific collection
func (kb *ChromaKB) Initialize() error {
	kb.collectionMu.Lock()
	defer kb.collectionMu.Unlock()
	fmt.Println("NewChromaKB Initialize")
	ctx := context.Background()
	//var collections []*chroma.Collection
	//var err error
	//collections, err = kb.client.ListCollections(ctx)
	//if err != nil {
	//	log.Printf("列出集合时出错: %v", err)
	//}
	//
	//if len(collections) > 0 {
	//	for _, c := range collections {
	//		if c.Name == kb.collectionName {
	//			kb.collection = c
	//			return nil
	//		}
	//	}
	//}

	metadata := map[string]interface{}{} // 使用空 map 或默认值
	embeddingFunction := types.NewConsistentHashEmbeddingFunction()
	fmt.Println("CreateCollection =>", kb.collectionName)
	newCollection, err := kb.client.CreateCollection(
		ctx,
		kb.collectionName,
		metadata,
		true,
		embeddingFunction,
		types.L2,
	)
	if err != nil {
		log.Fatalf("Error creating collection: %s \n", err)
	}
	kb.collection = newCollection
	return nil
}

// AddDocuments adds documents to the knowledge base
// AddDocuments adds documents to the knowledge base
func (kb *ChromaKB) AddDocuments(docs []Document) error {
	kb.collectionMu.Lock()
	defer kb.collectionMu.Unlock()

	ctx := context.Background()

	// Create a new record set with to hold the records to insert
	rs, err := types.NewRecordSet(
		types.WithEmbeddingFunction(kb.collection.EmbeddingFunction), // we pass the embedding function from the collection
		types.WithIDGenerator(types.NewULIDGenerator()),
	)
	if err != nil {
		log.Fatalf("Error creating record set: %s \n", err)
	}
	for _, doc := range docs {
		// Add a few records to the record set
		rs.WithRecord(types.WithDocument(doc.Text), types.WithMetadata("metadata", doc.Metadata))
	}
	// Build and validate the record set (this will create embeddings if not already present)
	_, err = rs.BuildAndValidate(ctx)
	if err != nil {
		log.Fatalf("Error validating record set: %s \n", err)
	}

	// Add the records to the collection
	_, err = kb.collection.AddRecords(context.Background(), rs)
	if err != nil {
		log.Fatalf("Error adding documents: %s \n", err)
	}

	return nil
}

// Query queries the knowledge base for documents
func (kb *ChromaKB) Query(query string, nResults int) ([]Document, error) {
	kb.collectionMu.RLock()
	defer kb.collectionMu.RUnlock()
	if kb.collection == nil {
		log.Printf("集合未初始化")
		return nil, fmt.Errorf("集合未初始化")
	}
	ctx := context.Background()
	//sl, _ := kb.embeddingFunc.EmbedQuery(ctx, query)
	//fmt.Println("EmbedQuery:", sl)
	fmt.Println("Query:", query, "  ", nResults)
	results, err := kb.collection.Query(ctx, []string{query}, int32(nResults), nil, nil, nil)
	if err != nil {
		log.Printf("查询文档时出错: %v", err)
		return nil, fmt.Errorf("查询文档时出错: %w", err)
	}
	fmt.Println("results:", results)

	var docs []Document
	for k, v := range results.Ids {
		ids := ""
		if len(v) > 0 {
			ids = v[0]
		}
		Text := ""
		if len(results.Documents) < k {
			Text = results.Documents[k][0]
		}
		var Metadatas map[string]interface{}
		if len(results.Metadatas) < k {
			Metadatas = results.Metadatas[k][0]
		}
		docs = append(docs, Document{
			ID:       ids,
			Text:     Text,
			Metadata: Metadatas,
		})
	}
	return docs, nil
}

// DeleteDocument deletes a document from the knowledge base
func (kb *ChromaKB) DeleteDocument(id string) error {
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
func (kb *ChromaKB) ListDocuments() ([]Document, error) {
	kb.collectionMu.RLock()
	defer kb.collectionMu.RUnlock()

	ctx := context.Background()
	results, err := kb.collection.Get(ctx, nil, nil, nil, nil) // 假设 Get 方法用于列出文档，具体方法名需根据库文档确认
	if err != nil {
		log.Printf("列出文档时出错: %v", err)
		return nil, fmt.Errorf("列出文档时出错: %w", err)
	}

	var docs []Document
	for k, v := range results.Ids {
		ids := v
		Text := ""
		if len(results.Documents) < k {
			Text = results.Documents[k]
		}
		var Metadatas map[string]interface{}
		if len(results.Metadatas) < k {
			Metadatas = results.Metadatas[k]
		}
		docs = append(docs, Document{
			ID:       ids,
			Text:     Text,
			Metadata: Metadatas,
		})
	}
	return docs, nil
}
