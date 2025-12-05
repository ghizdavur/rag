package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"cmd/main.go/pkg/rag"
)

func main() {
	mode := flag.String("mode", "ingest", "ingest or query")
	indexPath := flag.String("index", rag.DefaultIndexPath, "path to the rag index (JSON file)")
	docsDir := flag.String("docs", rag.DefaultLocalDocsFolder, "local docs directory to include during ingestion")
	chunkSize := flag.Int("chunk-size", 1400, "characters per chunk")
	chunkOverlap := flag.Int("chunk-overlap", 200, "character overlap between chunks")
	topK := flag.Int("top-k", rag.DefaultTopK, "number of chunks to send to the LLM in query mode")
	questionFlag := flag.String("question", "", "question to ask when mode=query")
	flag.Parse()

	ctx := context.Background()
	cfg := rag.LoadServiceConfigFromEnv()
	if cfg.Provider == rag.ProviderOpenAI && cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY must be set when RAG_PROVIDER=openai")
	}

	resolvedIndex := rag.ResolveWorkspacePath(*indexPath)

	switch strings.ToLower(*mode) {
	case "ingest":
		runIngest(ctx, cfg, rag.ResolveWorkspacePath(*docsDir), resolvedIndex, *chunkSize, *chunkOverlap)
	case "query":
		question := strings.TrimSpace(*questionFlag)
		if question == "" {
			question = strings.TrimSpace(strings.Join(flag.Args(), " "))
		}
		if question == "" {
			log.Fatal("provide a question via --question or as a positional argument, e.g. --mode query --question \"How do SP-API rate limits work?\"")
		}
		runQuery(ctx, cfg, question, resolvedIndex, *topK)
	default:
		log.Fatalf("unsupported mode %s", *mode)
	}
}

func runIngest(ctx context.Context, cfg rag.ServiceConfig, docsDir, indexPath string, chunkSize, chunkOverlap int) {
	opts := rag.DefaultSourceOptions(docsDir)
	documents, err := rag.CollectDocuments(ctx, opts)
	if err != nil {
		log.Fatalf("collect documents: %v", err)
	}
	if len(documents) == 0 {
		log.Fatal("no documents discovered for ingestion")
	}

	chunks := rag.ChunkDocuments(documents, rag.ChunkOptions{Size: chunkSize, Overlap: chunkOverlap})
	embedder, err := rag.NewEmbedder(cfg)
	if err != nil {
		log.Fatalf("create embedder: %v", err)
	}

	meta := rag.MetadataForRun(len(documents), len(chunks))
	store, err := rag.BuildVectorStore(ctx, chunks, embedder, 16, meta)
	if err != nil {
		log.Fatalf("build vector store: %v", err)
	}
	if err := store.Save(indexPath); err != nil {
		log.Fatalf("save vector store: %v", err)
	}

	fmt.Printf("Ingestion complete: %d documents -> %d chunks (saved at %s)\n", len(documents), len(chunks), indexPath)
}

func runQuery(ctx context.Context, cfg rag.ServiceConfig, question, indexPath string, topK int) {
	store, err := rag.LoadVectorStore(indexPath)
	if err != nil {
		log.Fatalf("load vector store: %v", err)
	}
	embedder, err := rag.NewEmbedder(cfg)
	if err != nil {
		log.Fatalf("create embedder: %v", err)
	}
	chatClient, err := rag.NewChatClient(cfg)
	if err != nil {
		log.Fatalf("create chat client: %v", err)
	}

	service := rag.NewService(store, embedder, chatClient, cfg)
	answer, err := service.Answer(ctx, question, rag.QueryOptions{TopK: topK})
	if err != nil {
		log.Fatalf("query rag: %v", err)
	}

	fmt.Println("Answer:\n", answer.Answer)
	fmt.Println("\nSources:")
	for _, src := range answer.Sources {
		fmt.Printf("- (%.3f) %s => %s\n", src.Score, src.Title, src.URI)
	}
}
