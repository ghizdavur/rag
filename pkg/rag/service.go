package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Service wires the vector store, embedder, and LLM together.
type Service struct {
	store        *VectorStore
	embedder     Embedder
	chatClient   ChatClient
	systemPrompt string
	defaultTopK  int
}

// NewService creates a ready-to-use RAG service.
func NewService(store *VectorStore, embedder Embedder, chatClient ChatClient, cfg ServiceConfig) *Service {
	topK := cfg.DefaultTopK
	if topK <= 0 {
		topK = DefaultTopK
	}
	prompt := cfg.SystemPrompt
	if prompt == "" {
		prompt = DefaultSystemPrompt
	}
	return &Service{
		store:        store,
		embedder:     embedder,
		chatClient:   chatClient,
		systemPrompt: prompt,
		defaultTopK:  topK,
	}
}

// NewServiceFromEnv loads configuration and supporting assets from disk.
func NewServiceFromEnv(ctx context.Context) (*Service, error) {
	cfg := LoadServiceConfigFromEnv()
	store, err := LoadVectorStore(cfg.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("load vector store: %w", err)
	}
	embedder, err := NewEmbedder(cfg)
	if err != nil {
		return nil, err
	}
	chatClient, err := NewChatClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewService(store, embedder, chatClient, cfg), nil
}

// Answer runs retrieval + generation.
func (s *Service) Answer(ctx context.Context, question string, opts QueryOptions) (*Answer, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("rag service is not initialized")
	}
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return nil, errors.New("question is required")
	}
	if opts.TopK <= 0 {
		opts.TopK = s.defaultTopK
	}
	if opts.Temperature == 0 {
		opts.Temperature = 0.2
	}

	// Step 1: Generate query embedding
	embedStart := time.Now()
	embeddings, err := s.embedder.Embed(ctx, []string{trimmed})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("empty query embedding")
	}
	embedDuration := time.Since(embedStart)
	fmt.Printf("[PERF] Query embedding: %v\n", embedDuration)

	// Step 2: Search vector store
	searchStart := time.Now()
	matches := s.store.Search(embeddings[0], opts.TopK)
	if len(matches) == 0 {
		return nil, errors.New("no context available; run ingestion first")
	}
	searchDuration := time.Since(searchStart)
	fmt.Printf("[PERF] Vector search (%d chunks): %v\n", len(s.store.Chunks), searchDuration)

	// Step 3: Generate answer
	prompt := buildPrompt(trimmed, matches)
	genStart := time.Now()
	answer, err := s.chatClient.Complete(ctx, s.systemPrompt, prompt, opts.Temperature)
	genDuration := time.Since(genStart)
	fmt.Printf("[PERF] LLM generation: %v\n", genDuration)
	fmt.Printf("[PERF] Total time: %v\n", time.Since(embedStart))
	if err != nil {
		return nil, err
	}

	attributions := make([]SourceAttribution, len(matches))
	for i, match := range matches {
		snippet := strings.TrimSpace(match.Chunk.Text)
		if len(snippet) > 400 {
			snippet = snippet[:400] + "..."
		}
		attributions[i] = SourceAttribution{
			Title:   match.Chunk.Source,
			URI:     match.Chunk.URI,
			Snippet: snippet,
			Score:   match.Score,
		}
	}

	return &Answer{Answer: strings.TrimSpace(answer), Sources: attributions}, nil
}

func buildPrompt(question string, matches []SearchResult) string {
	var b strings.Builder
	b.WriteString("Context sections (most relevant to least):\n")
	for i, match := range matches {
		b.WriteString(fmt.Sprintf("[%d] Source: %s (%s)\n", i+1, match.Chunk.Source, match.Chunk.URI))
		b.WriteString(match.Chunk.Text)
		b.WriteString("\n\n")
	}

	b.WriteString("Instructions:\n")
	b.WriteString("1. Use only the provided context sections.\n")
	b.WriteString("2. If the answer is not present, say you do not have that information.\n")
	b.WriteString("3. When relevant, cite the source title in parentheses.\n")
	b.WriteString("4. Highlight Amazon-specific constraints (rate limits, launch phases, pilots) explicitly.\n")

	b.WriteString("\nQuestion:\n")
	b.WriteString(question)

	return b.String()
}

// MetadataForRun captures metadata for ingestion runs.
func MetadataForRun(sourceCount, chunkCount int) Metadata {
	return Metadata{
		GeneratedAt: time.Now().UTC(),
		SourceCount: sourceCount,
		ChunkCount:  chunkCount,
	}
}

// AddSource adds a new text source to the existing vector store.
func (s *Service) AddSource(ctx context.Context, title, content, uri string) error {
	if s == nil || s.store == nil {
		return errors.New("rag service is not initialized")
	}
	if content == "" {
		return errors.New("content cannot be empty")
	}
	if title == "" {
		title = "User Added Source"
	}
	if uri == "" {
		uri = "user-input://" + title
	}

	// Create document from text
	doc := Document{
		ID:      Slugify(title),
		Title:   title,
		URI:     uri,
		Source:  "user-added",
		Content: strings.TrimSpace(content),
	}

	// Chunk the document
	chunks := ChunkDocuments([]Document{doc}, ChunkOptions{Size: 1400, Overlap: 200})

	// Embed chunks
	embedder := s.embedder
	for i := range chunks {
		texts := []string{chunks[i].Text}
		embeddings, err := embedder.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("failed to embed chunk: %w", err)
		}
		if len(embeddings) > 0 {
			chunks[i].Embedding = embeddings[0]
		}
		// Small delay to avoid overwhelming Ollama
		time.Sleep(500 * time.Millisecond)
	}

	// Add chunks to existing store
	s.store.Chunks = append(s.store.Chunks, chunks...)
	s.store.Metadata.SourceCount++
	s.store.Metadata.ChunkCount += len(chunks)
	s.store.Metadata.GeneratedAt = time.Now().UTC()

	return nil
}

// SaveStore saves the current vector store to disk.
func (s *Service) SaveStore(indexPath string) error {
	if s == nil || s.store == nil {
		return errors.New("rag service is not initialized")
	}
	return s.store.Save(indexPath)
}
