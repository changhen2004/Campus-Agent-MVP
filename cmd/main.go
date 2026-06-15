package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"campus-agent/internal/ai/client"
	"campus-agent/internal/ai/embedder"
	"campus-agent/internal/ai/retriever"
	"campus-agent/internal/handler"
	localknowledge "campus-agent/internal/knowledge/local"
	"campus-agent/internal/repo/qdrant"
	"campus-agent/internal/router"
	"campus-agent/internal/server/chat"
	"campus-agent/internal/server/knowledge"
	knowledgetool "campus-agent/internal/tool/knowledge"
	"campus-agent/pkg/config"
)

func main() {
	configPath := os.Getenv("CAMPUS_AGENT_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	// Load local knowledge docs as fallback
	knowledgeDocs, err := localknowledge.LoadMarkdownDir("docs/knowledge")
	if err != nil {
		log.Printf("[main] load knowledge docs: %v", err)
		knowledgeDocs = []knowledgetool.Document{}
	}
	log.Printf("[main] loaded %d knowledge documents from docs/knowledge", len(knowledgeDocs))

	// Initialize local knowledge store (always available as fallback)
	localStore := knowledgetool.NewLocalTool(knowledgeDocs)

	// Initialize LLM client
	llmClient := client.New(cfg.LLM)

	// Attempt to initialize Qdrant + Embedding
	var qdrantClient *qdrant.Client
	var qdrantIdx *qdrant.Indexer
	var qdrantRet *qdrant.Retriever
	var embedderService *embedder.Embedder

	if cfg.Embedding.APIKey != "" && cfg.Embedding.APIKey != "replace-me" {
		embedderService = embedder.New(cfg.Embedding)

		qdrantClient, err = qdrant.NewClient(ctx, cfg.Qdrant, cfg.Embedding.Dimension)
		if err != nil {
			log.Printf("[main] qdrant init failed (will use local search): %v", err)
		} else {
			defer qdrantClient.Close()

			qdrantIdx = qdrant.NewIndexer(qdrantClient, embedderService)
			qdrantRet = qdrant.NewRetriever(qdrantClient, embedderService)

			// Index existing documents into Qdrant
			if len(knowledgeDocs) > 0 {
				docs := make([]qdrant.Document, 0, len(knowledgeDocs))
				for _, d := range knowledgeDocs {
					docs = append(docs, qdrant.Document{ID: d.ID, Title: d.Title, Content: d.Content})
				}
				if err := qdrantIdx.IndexDocuments(ctx, docs); err != nil {
					log.Printf("[main] index existing docs to qdrant: %v", err)
				}
			}
		}
	}

	// Initialize retriever (Qdrant + local fallback)
	ret := retriever.New(qdrantRet, localStore, cfg.RAG)

	// Initialize services
	chatService := chat.NewService(llmClient, ret, cfg.RAG)
	knowledgeService := knowledge.NewService(localStore, qdrantIdx)

	// Initialize handlers
	chatHandler := handler.NewChatHandler(chatService)
	knowledgeHandler := handler.NewKnowledgeHandler(knowledgeService)
	staticFS := http.FS(os.DirFS("web/static"))

	// Setup router
	r := router.New(chatHandler, knowledgeHandler, staticFS)

	addr := cfg.Server.Addr()
	log.Printf("[main] campus-agent server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
