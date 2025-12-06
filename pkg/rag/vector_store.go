package rag

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
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

		// Retry logic for Ollama connection issues on Windows
		var embeddings [][]float32
		var err error
		maxRetries := 5 // Increased retries
		for attempt := 0; attempt < maxRetries; attempt++ {
			embeddings, err = embedder.Embed(ctx, texts)
			if err == nil {
				break
			}
			if attempt < maxRetries-1 {
				backoff := time.Duration(attempt+1) * 1 * time.Second // Increased backoff
				time.Sleep(backoff)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to embed batch [%d:%d] after %d attempts: %w", start, end, maxRetries, err)
		}

		for i := range batch {
			chunks[start+i].Embedding = embeddings[i]
		}
		// Add longer delay between batches to avoid overwhelming Ollama on Windows
		if start+batchSize < len(chunks) {
			time.Sleep(1 * time.Second) // Increased to 1 second
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
// Optimized to use a min-heap for better performance with large vector stores.
func (vs *VectorStore) Search(query []float32, topK int) []SearchResult {
	if vs == nil || len(query) == 0 {
		return nil
	}
	if topK <= 0 {
		topK = 4
	}
	if len(vs.Chunks) == 0 {
		return nil
	}

	// Use min-heap to maintain only topK results (more efficient than sorting all)
	pq := make(PriorityQueue, 0, topK+1)
	heap.Init(&pq)

	for _, chunk := range vs.Chunks {
		score := cosineSimilarity(query, chunk.Embedding)

		// If heap is not full, add the result
		if pq.Len() < topK {
			heap.Push(&pq, &Item{
				chunk: chunk,
				score: score,
			})
		} else {
			// If heap is full, only add if score is better than the worst in heap
			worst := pq[0]
			if score > worst.score {
				heap.Pop(&pq)
				heap.Push(&pq, &Item{
					chunk: chunk,
					score: score,
				})
			}
		}
	}

	// Extract results from heap and sort by score (descending)
	results := make([]SearchResult, pq.Len())
	for i := pq.Len() - 1; i >= 0; i-- {
		item := heap.Pop(&pq).(*Item)
		results[i] = SearchResult{
			Chunk: item.chunk,
			Score: item.score,
		}
	}

	// Reverse to get descending order (highest score first)
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	// Optimized cosine similarity calculation
	// Pre-compute magnitudes if possible, but for now use optimized loop
	var dot float64
	var magA float64
	var magB float64

	// Unroll loop for better performance (process 4 at a time)
	n := len(a)
	i := 0
	for i < n-3 {
		dot += float64(a[i]*b[i] + a[i+1]*b[i+1] + a[i+2]*b[i+2] + a[i+3]*b[i+3])
		magA += float64(a[i]*a[i] + a[i+1]*a[i+1] + a[i+2]*a[i+2] + a[i+3]*a[i+3])
		magB += float64(b[i]*b[i] + b[i+1]*b[i+1] + b[i+2]*b[i+2] + b[i+3]*b[i+3])
		i += 4
	}
	// Handle remaining elements
	for i < n {
		dot += float64(a[i] * b[i])
		magA += float64(a[i] * a[i])
		magB += float64(b[i] * b[i])
		i++
	}

	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

// PriorityQueue implements a min-heap for efficient topK search
type Item struct {
	chunk Chunk
	score float64
	index int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Min-heap: lower score has higher priority (we want to keep highest scores)
	return pq[i].score < pq[j].score
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
