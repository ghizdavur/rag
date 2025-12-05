package rag

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// VectorStore persists embedded chunks on disk for later querying.
type VectorStore struct {
	Metadata Metadata `json:"metadata"`
	Chunks   []Chunk  `json:"chunks"`
}

// BuildVectorStore embeds all chunks and returns a ready-to-save store.
func BuildVectorStore(ctx context.Context, chunks []Chunk, embedder Embedder, batchSize int, meta Metadata) (*VectorStore, error) {
	if embedder == nil {
		return nil, errors.New("embedder is required")
	}
	if len(chunks) == 0 {
		return nil, errors.New("no chunks supplied")
	}
	if batchSize <= 0 {
		batchSize = 16
	}

	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[start:end]
		texts := make([]string, len(batch))
		for i, chunk := range batch {
			texts[i] = chunk.Text
		}
		embeddings, err := embedder.Embed(ctx, texts)
		if err != nil {
			return nil, err
		}
		for i := range batch {
			chunks[start+i].Embedding = embeddings[i]
		}
	}

	store := &VectorStore{Metadata: meta, Chunks: chunks}
	return store, nil
}

// Save writes the vector store to disk.
func (vs *VectorStore) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(vs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadVectorStore reads a store from disk.
func LoadVectorStore(path string) (*VectorStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var store VectorStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

// SearchResult describes the best-matching chunks.
type SearchResult struct {
	Chunk Chunk
	Score float64
}

// Search returns the topK chunks that best match the supplied embedding.
func (vs *VectorStore) Search(query []float32, topK int) []SearchResult {
	if vs == nil || len(query) == 0 {
		return nil
	}
	if topK <= 0 {
		topK = 4
	}
	results := make([]SearchResult, 0, topK)
	for _, chunk := range vs.Chunks {
		score := cosineSimilarity(query, chunk.Embedding)
		results = append(results, SearchResult{Chunk: chunk, Score: score})
	}
	sortByScore(results)
	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot float64
	var magA float64
	var magB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		magA += float64(a[i] * a[i])
		magB += float64(b[i] * b[i])
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

func sortByScore(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
