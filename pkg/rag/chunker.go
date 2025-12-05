package rag

import (
	"fmt"
	"unicode/utf8"
)

// ChunkOptions controls how large each chunk becomes.
type ChunkOptions struct {
	Size    int
	Overlap int
}

// ChunkDocuments splits documents into overlapping windows for embedding.
func ChunkDocuments(docs []Document, opts ChunkOptions) []Chunk {
	if opts.Size <= 0 {
		opts.Size = 1200
	}
	if opts.Overlap < 0 {
		opts.Overlap = 0
	}
	if opts.Overlap >= opts.Size {
		opts.Overlap = opts.Size / 4
	}

	chunks := make([]Chunk, 0, len(docs)*4)

	for _, doc := range docs {
		windows := slidingWindows(doc.Content, opts.Size, opts.Overlap)
		for idx, text := range windows {
			chunkID := fmt.Sprintf("%s-chunk-%d", doc.ID, idx)
			chunks = append(chunks, Chunk{
				ID:         chunkID,
				DocumentID: doc.ID,
				Source:     doc.Title,
				URI:        doc.URI,
				Text:       text,
				Index:      idx,
			})
		}
	}

	return chunks
}

func slidingWindows(content string, size, overlap int) []string {
	runeCount := utf8.RuneCountInString(content)
	if runeCount == 0 {
		return nil
	}

	if runeCount <= size {
		return []string{content}
	}

	step := size - overlap
	if step <= 0 {
		step = size
	}

	windows := []string{}
	runes := []rune(content)
	for start := 0; start < len(runes); start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		window := runes[start:end]
		windows = append(windows, string(window))
		if end == len(runes) {
			break
		}
	}
	return windows
}
