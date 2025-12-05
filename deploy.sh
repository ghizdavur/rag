#!/bin/bash

# RAG Application Deployment Script
# Usage: ./deploy.sh

set -e

echo "🚀 Starting RAG Application Deployment..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Go is installed${NC}"

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo -e "${YELLOW}⚠️  Ollama is not running. Starting Ollama...${NC}"
    sudo systemctl start ollama || echo -e "${RED}❌ Failed to start Ollama${NC}"
    sleep 3
fi

# Check if Ollama models are available
if ! ollama list | grep -q "llama3:8b"; then
    echo -e "${YELLOW}⚠️  llama3:8b model not found. Pulling...${NC}"
    ollama pull llama3:8b
fi

if ! ollama list | grep -q "nomic-embed-text"; then
    echo -e "${YELLOW}⚠️  nomic-embed-text model not found. Pulling...${NC}"
    ollama pull nomic-embed-text:latest
fi

echo -e "${GREEN}✅ Ollama is ready${NC}"

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download

# Build application
echo "🔨 Building application..."
go build -ldflags="-s -w" -o rag-server ./cmd/main.go

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found. Creating template...${NC}"
    cat > .env << EOF
# RAG Configuration
RAG_PROVIDER=ollama
RAG_INDEX_PATH=$(pwd)/data/rag_index.json
RAG_OLLAMA_BASE_URL=http://localhost:11434
RAG_EMBEDDING_MODEL=nomic-embed-text:latest
RAG_CHAT_MODEL=llama3:8b
EOF
    echo -e "${YELLOW}⚠️  Please edit .env file with your configuration${NC}"
fi

# Create data directory
mkdir -p data
mkdir -p docs

# Check if index exists, if not run ingestion
if [ ! -f data/rag_index.json ]; then
    echo "📚 Running initial ingestion..."
    go run ./cmd/rag --mode ingest --index data/rag_index.json
fi

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""
echo "To run the application:"
echo "  ./rag-server"
echo ""
echo "Or with systemd:"
echo "  sudo systemctl start rag"
echo ""

