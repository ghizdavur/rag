# How to Add More Sources to RAG

## Method 1: Add Local Documents (Easiest)

Simply add files to the `docs/` folder:

### Supported formats:
- `.md` (Markdown)
- `.markdown` (Markdown)
- `.txt` (Plain text)

### Steps:
1. Add your files to `docs/` folder
2. Run ingestion: `go run ./cmd/rag --mode ingest --index data/rag_index.json`
3. Restart the server

### Example:
```bash
# Add a new document
echo "# My Documentation" > docs/my-docs.md

# Run ingestion
go run ./cmd/rag --mode ingest --index data/rag_index.json

# Restart server
go run cmd/main.go
```

## Method 2: Add Remote Sources (URLs)

Edit `pkg/rag/documents.go` and add remote sources to `DefaultSourceOptions`:

```go
RemoteSources: []RemoteSource{
    {
        Name:        "Your Source Name",
        URL:         "https://example.com/documentation",
        Format:      FormatMarkdown, // or FormatHTML, FormatText, FormatTSV
        Description: "Description of the source",
    },
}
```

### Supported formats:
- `FormatMarkdown` - For .md files
- `FormatHTML` - For web pages (will be converted to text)
- `FormatText` - For plain text
- `FormatTSV` - For tab-separated values

## Method 3: Use Different Document Folder

You can point to a different folder:

```bash
go run ./cmd/rag --mode ingest --index data/rag_index.json --docs /path/to/your/docs
```

## After Adding Sources

Always run ingestion after adding new sources:

```bash
go run ./cmd/rag --mode ingest --index data/rag_index.json
```

Then restart the server to use the new index.

