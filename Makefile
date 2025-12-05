.PHONY: build run deploy install clean test

# Build the application
build:
	@echo "Building RAG application..."
	go build -ldflags="-s -w" -o bin/rag-server ./cmd/main.go

# Run the application
run:
	@echo "Running RAG application..."
	go run cmd/main.go

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run ingestion
ingest:
	@echo "Running ingestion..."
	go run ./cmd/rag --mode ingest --index data/rag_index.json

# Deploy to server (requires SSH access)
deploy:
	@echo "Deploying to server..."
	git push origin main
	# Add your deployment commands here

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f bin/rag-server
	rm -f rag-server
	rm -f *.exe

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Build for production
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/rag-server-linux ./cmd/main.go

# Setup (first time)
setup:
	@echo "Setting up RAG application..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file..."; \
		echo "RAG_PROVIDER=ollama" > .env; \
		echo "RAG_INDEX_PATH=$(PWD)/data/rag_index.json" >> .env; \
	fi
	@mkdir -p data docs
	@echo "✅ Setup complete!"
	@echo "Edit .env file and run 'make ingest' to create the index"

