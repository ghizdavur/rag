# Example: How to Add Remote Sources

To add remote sources (URLs), edit `pkg/rag/documents.go`:

## Example 1: Add a GitHub README

```go
RemoteSources: []RemoteSource{
    {
        Name:        "My Project README",
        URL:         "https://raw.githubusercontent.com/user/repo/main/README.md",
        Format:      FormatMarkdown,
        Description: "Main project documentation",
    },
}
```

## Example 2: Add a Documentation Website

```go
RemoteSources: []RemoteSource{
    {
        Name:        "API Documentation",
        URL:         "https://example.com/api/docs",
        Format:      FormatHTML, // Will convert HTML to text
        Description: "Official API documentation",
    },
}
```

## Example 3: Add Multiple Sources

```go
RemoteSources: []RemoteSource{
    {
        Name:        "Source 1",
        URL:         "https://example.com/doc1.md",
        Format:      FormatMarkdown,
        Description: "First source",
    },
    {
        Name:        "Source 2",
        URL:         "https://example.com/doc2.html",
        Format:      FormatHTML,
        Description: "Second source",
    },
}
```

After editing, run ingestion:
```bash
go run ./cmd/rag --mode ingest --index data/rag_index.json
```


