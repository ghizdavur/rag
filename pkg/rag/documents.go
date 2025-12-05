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
		RemoteSources: []RemoteSource{
			{
				Name:        "Amazon Selling Partner API Samples (README)",
				URL:         "https://raw.githubusercontent.com/amzn/selling-partner-api-samples/main/README.md",
				Format:      FormatMarkdown,
				Description: "GitHub samples that showcase core SP-API workflows",
			},
			{
				Name:        "Selling Partner API Rate Limit Guide",
				URL:         "https://developer-docs.amazon.com/sp-api/docs/optimize-calls-to-the-selling-partner-api?ld=ASXXSPAPIDirect&pageName=US%3ASPDS%3ASPAPI-fees",
				Format:      FormatHTML,
				Description: "Amazon's official guidance on optimizing Selling Partner API usage",
			},
			{
				Name:        "Selling Partner API Documentation Portal",
				URL:         "https://developer-docs.amazon.com/sp-api",
				Format:      FormatHTML,
				Description: "Landing site for all SP-API documentation",
			},
			{
				Name:        "Amazon Pilot + Feature Toggle Tracker",
				URL:         "https://docs.google.com/spreadsheets/d/1L0AkVtKDOuvYLkeHbYY9McJxcgyVxFJqDdbXsAmmFaM/export?format=tsv&gid=0",
				Format:      FormatTSV,
				Description: "Internal sheet with pilot customers and beta configurations",
			},
			{
				Name:        "plentymarkets Amazon MC repositories",
				URL:         "https://github.com/orgs/plentymarkets/repositories?language=&q=mc-amazon&sort=&type=all",
				Format:      FormatHTML,
				Description: "Partner-maintained repos that integrate with Amazon",
			},
		},
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
			ID:      slugify(rel),
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
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", src.URL, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", src.URL, err)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("fetch %s: status %d", src.URL, resp.StatusCode)
		}

		text, err := convertPayload(string(body), src.Format)
		if err != nil {
			return nil, fmt.Errorf("convert %s: %w", src.URL, err)
		}

		documents = append(documents, Document{
			ID:      slugify(src.Name),
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

func slugify(input string) string {
	lower := strings.ToLower(input)
	slug := slugMatcher.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "doc"
	}
	return slug
}
