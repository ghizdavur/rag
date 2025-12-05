package rag

import (
	"os"
	"path/filepath"
	"strconv"
)

const (
	// DefaultIndexPath points to the generated vector store relative to the repository root.
	DefaultIndexPath       = "data/rag_index.json"
	DefaultEmbeddingModel  = "text-embedding-3-large"
	DefaultChatModel       = "gpt-4o-mini"
	DefaultSystemPrompt    = "You are an assistant that answers questions about Amazon Selling Partner integrations. Reply with concise, implementation-focused answers and cite the provided context snippets."
	DefaultTopK            = 4
	DefaultLocalDocsFolder = "docs"
)

// ServiceConfig controls how the runtime RAG service behaves.
type ServiceConfig struct {
	IndexPath      string
	OpenAIAPIKey   string
	EmbeddingModel string
	ChatModel      string
	SystemPrompt   string
	DefaultTopK    int
}

// LoadServiceConfigFromEnv loads runtime RAG configuration from environment variables.
func LoadServiceConfigFromEnv() ServiceConfig {
	indexPath := firstNonEmpty(os.Getenv("RAG_INDEX_PATH"), DefaultIndexPath)
	embeddingModel := firstNonEmpty(os.Getenv("RAG_EMBEDDING_MODEL"), DefaultEmbeddingModel)
	chatModel := firstNonEmpty(os.Getenv("RAG_CHAT_MODEL"), DefaultChatModel)
	systemPrompt := firstNonEmpty(os.Getenv("RAG_SYSTEM_PROMPT"), DefaultSystemPrompt)
	topK := parseIntEnv("RAG_DEFAULT_TOP_K", DefaultTopK)

	return ServiceConfig{
		IndexPath:      resolveWorkspacePath(indexPath),
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		EmbeddingModel: embeddingModel,
		ChatModel:      chatModel,
		SystemPrompt:   systemPrompt,
		DefaultTopK:    topK,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseIntEnv(key string, fallback int) int {
	if raw := os.Getenv(key); raw != "" {
		if val, err := strconv.Atoi(raw); err == nil {
			return val
		}
	}
	return fallback
}

// ResolveWorkspacePath exposes the internal helper for other packages, e.g. CLI tooling.
func ResolveWorkspacePath(pathValue string) string {
	return resolveWorkspacePath(pathValue)
}

func resolveWorkspacePath(pathValue string) string {
	if pathValue == "" {
		return pathValue
	}
	if filepath.IsAbs(pathValue) {
		return filepath.Clean(pathValue)
	}

	if root := findRepoRoot(); root != "" {
		return filepath.Join(root, pathValue)
	}

	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, pathValue)
	}
	return filepath.Clean(pathValue)
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(wd, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return ""
}
