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
	if cfg.OpenAIAPIKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not configured")
	}
	store, err := LoadVectorStore(cfg.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("load vector store: %w", err)
	}
	embedder, err := NewOpenAIEmbedder(cfg.OpenAIAPIKey, cfg.EmbeddingModel)
	if err != nil {
		return nil, err
	}
	chatClient, err := NewOpenAIChatClient(cfg.OpenAIAPIKey, cfg.ChatModel)
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

	embeddings, err := s.embedder.Embed(ctx, []string{trimmed})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, errors.New("empty query embedding")
	}

	matches := s.store.Search(embeddings[0], opts.TopK)
	if len(matches) == 0 {
		return nil, errors.New("no context available; run ingestion first")
	}

	prompt := buildPrompt(trimmed, matches)
	answer, err := s.chatClient.Complete(ctx, s.systemPrompt, prompt, opts.Temperature)
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
