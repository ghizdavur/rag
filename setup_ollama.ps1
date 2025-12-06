# Script pentru setup Ollama
Write-Host "Instalare modele Ollama pentru RAG..." -ForegroundColor Green

Write-Host "`n1. Pulling embedding model (nomic-embed-text)..." -ForegroundColor Yellow
ollama pull nomic-embed-text

Write-Host "`n2. Pulling chat model (llama3:8b)..." -ForegroundColor Yellow
ollama pull llama3:8b

Write-Host "`n✅ Setup complet! Acum poți rula ingestion-ul." -ForegroundColor Green
Write-Host "Rulează: go run ./cmd/rag --mode ingest --index data/rag_index.json" -ForegroundColor Cyan


