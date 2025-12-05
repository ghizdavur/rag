package rag

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jaytaylor/html2text"
)

// RemoteFormat enumerates the strategies for parsing downloaded content.
type RemoteFormat string

const (
	FormatMarkdown RemoteFormat = "markdown"
	FormatHTML     RemoteFormat = "html"
	FormatText     RemoteFormat = "text"
	FormatTSV      RemoteFormat = "tsv"
)

// RemoteSource declares a remote artifact to ingest.
type RemoteSource struct {
	Name        string
	URL         string
	Format      RemoteFormat
	Description string
}

// SourceOptions controls how we discover documents.
type SourceOptions struct {
	LocalDocsDir      string
	IncludeExtensions []string
	RemoteSources     []RemoteSource
}

// DefaultSourceOptions returns a pre-populated list using the resources shared by the team.
func DefaultSourceOptions(baseDir string) SourceOptions {
	if baseDir == "" {
		baseDir = DefaultLocalDocsFolder
	}
	baseDir = resolveWorkspacePath(baseDir)

	return SourceOptions{
		LocalDocsDir:      baseDir,
		IncludeExtensions: []string{".md", ".markdown", ".txt"},
		RemoteSources:     []RemoteSource{}, // Disabled remote sources to avoid Ollama issues on Windows
	}
}

// CollectDocuments walks both local and remote sources.
func CollectDocuments(ctx context.Context, opts SourceOptions) ([]Document, error) {
	var documents []Document

	if localDocs, err := collectLocalDocuments(opts); err == nil {
		documents = append(documents, localDocs...)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("collect local docs: %w", err)
	}

	if len(opts.RemoteSources) > 0 {
		remoteDocs, err := collectRemoteDocuments(ctx, opts.RemoteSources)
		if err != nil {
			return nil, fmt.Errorf("collect remote docs: %w", err)
		}
		documents = append(documents, remoteDocs...)
	}

	return documents, nil
}

func collectLocalDocuments(opts SourceOptions) ([]Document, error) {
	info, err := os.Stat(opts.LocalDocsDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", opts.LocalDocsDir)
	}

	var documents []Document
	allowed := map[string]struct{}{}
	for _, ext := range opts.IncludeExtensions {
		allowed[strings.ToLower(ext)] = struct{}{}
	}

	err = filepath.WalkDir(opts.LocalDocsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if _, ok := allowed[strings.ToLower(filepath.Ext(entry.Name()))]; !ok {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(opts.LocalDocsDir, path)
		content := normalizeWhitespace(string(data))
		documents = append(documents, Document{
			ID:      Slugify(rel),
			Title:   fmt.Sprintf("Local: %s", rel),
			URI:     path,
			Source:  "local-docs",
			Content: content,
		})
		return nil
	})

	return documents, err
}

func collectRemoteDocuments(ctx context.Context, sources []RemoteSource) ([]Document, error) {
	client := &http.Client{Timeout: 45 * time.Second}
	documents := make([]Document, 0, len(sources))
	for _, src := range sources {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
		if err != nil {
			fmt.Printf("Warning: failed to create request for %s: %v\n", src.Name, err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Warning: failed to fetch %s: %v\n", src.Name, err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Warning: failed to read response for %s: %v\n", src.Name, err)
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			fmt.Printf("Warning: %s returned status %d, skipping\n", src.Name, resp.StatusCode)
			continue
		}

		text, err := convertPayload(string(body), src.Format)
		if err != nil {
			fmt.Printf("Warning: failed to convert %s: %v\n", src.Name, err)
			continue
		}

		documents = append(documents, Document{
			ID:      Slugify(src.Name),
			Title:   src.Name,
			URI:     src.URL,
			Source:  src.Description,
			Content: text,
		})
	}
	return documents, nil
}

func convertPayload(raw string, format RemoteFormat) (string, error) {
	switch format {
	case FormatMarkdown, FormatText, FormatTSV:
		return normalizeWhitespace(raw), nil
	case FormatHTML:
		text, err := html2text.FromString(raw, html2text.Options{PrettyTables: true})
		if err != nil {
			return "", err
		}
		return normalizeWhitespace(text), nil
	default:
		return "", fmt.Errorf("unsupported format %s", format)
	}
}

func normalizeWhitespace(input string) string {
	cleaned := strings.ReplaceAll(input, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	lines := strings.Split(cleaned, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		trimmed = append(trimmed, line)
	}
	return strings.Join(trimmed, "\n")
}

var slugMatcher = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a string to a URL-friendly slug.
func Slugify(input string) string {
	lower := strings.ToLower(input)
	slug := slugMatcher.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "doc"
	}
	return slug
}
