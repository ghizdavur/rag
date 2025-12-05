package rag

import "time"

// Document represents a normalized text artifact that will be chunked into embeddings.
type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URI     string `json:"uri"`
	Source  string `json:"source"`
	Content string `json:"content"`
}

// Chunk represents a slice of a document used for retrieval.
type Chunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	Source     string    `json:"source"`
	URI        string    `json:"uri"`
	Text       string    `json:"text"`
	Index      int       `json:"index"`
	Embedding  []float32 `json:"embedding"`
}

// Metadata tracks ingestion run details.
type Metadata struct {
	GeneratedAt time.Time `json:"generatedAt"`
	SourceCount int       `json:"sourceCount"`
	ChunkCount  int       `json:"chunkCount"`
	Notes       []string  `json:"notes"`
}

// QueryOptions configure retrieval and generation.
type QueryOptions struct {
	TopK        int
	Temperature float32
}

// Answer bundles the LLM output and retrieved snippets.
type Answer struct {
	Answer  string              `json:"answer"`
	Sources []SourceAttribution `json:"sources"`
}

// SourceAttribution highlights which slices backed the answer.
type SourceAttribution struct {
	Title   string  `json:"title"`
	URI     string  `json:"uri"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}
