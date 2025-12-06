package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Embedder converts text into vector representations.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// ChatClient generates answers from context-augmented prompts.
type ChatClient interface {
	Complete(ctx context.Context, systemPrompt, prompt string, temperature float32) (string, error)
}

// NewEmbedder returns an embedder based on the configured provider.
func NewEmbedder(cfg ServiceConfig) (Embedder, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return NewOllamaEmbedder(cfg.OllamaBaseURL, cfg.EmbeddingModel)
	case ProviderOpenAI:
		return NewOpenAIEmbedder(cfg.OpenAIAPIKey, cfg.EmbeddingModel)
	default:
		return nil, fmt.Errorf("unsupported provider %s", cfg.Provider)
	}
}

// NewChatClient returns a chat client for the configured provider.
func NewChatClient(cfg ServiceConfig) (ChatClient, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return NewOllamaChatClient(cfg.OllamaBaseURL, cfg.ChatModel), nil
	case ProviderOpenAI:
		return NewOpenAIChatClient(cfg.OpenAIAPIKey, cfg.ChatModel)
	default:
		return nil, fmt.Errorf("unsupported provider %s", cfg.Provider)
	}
}

// OpenAIEmbedder implements Embedder using the OpenAI embeddings API.
type OpenAIEmbedder struct {
	client *openai.Client
	model  string
}

// NewOpenAIEmbedder constructs an embedder for the supplied model.
func NewOpenAIEmbedder(apiKey, model string) (*OpenAIEmbedder, error) {
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}
	if model == "" {
		model = DefaultOpenAIEmbeddingModel
	}
	cfg := openai.DefaultConfig(apiKey)
	return &OpenAIEmbedder{client: openai.NewClientWithConfig(cfg), model: model}, nil
}

// Embed converts one or more texts into embedding vectors.
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	req := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(e.model),
		Input: texts,
	}
	resp, err := e.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}
	return embeddings, nil
}

// OpenAIChatClient implements ChatClient using the Chat Completions API.
type OpenAIChatClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIChatClient creates a chat completion client.
func NewOpenAIChatClient(apiKey, model string) (*OpenAIChatClient, error) {
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is required")
	}
	if model == "" {
		model = DefaultOpenAIChatModel
	}
	cfg := openai.DefaultConfig(apiKey)
	return &OpenAIChatClient{client: openai.NewClientWithConfig(cfg), model: model}, nil
}

// Complete generates an answer using the provided prompt.
func (c *OpenAIChatClient) Complete(ctx context.Context, systemPrompt, prompt string, temperature float32) (string, error) {
	if temperature == 0 {
		temperature = 0.2
	}
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: temperature,
		MaxTokens:   800,
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no chat completion choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}

// OllamaEmbedder implements Embedder using a local Ollama instance.
type OllamaEmbedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaEmbedder constructs an embedder backed by Ollama's /api/embed endpoint.
func NewOllamaEmbedder(baseURL, model string) (*OllamaEmbedder, error) {
	if model == "" {
		model = DefaultOllamaEmbeddingModel
	}
	if baseURL == "" {
		baseURL = DefaultOllamaBaseURL
	}
	return &OllamaEmbedder{
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	payload := map[string]interface{}{
		"model": e.model,
		"input": texts,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var parsed struct {
		Embeddings [][]float64 `json:"embeddings"`
		Embedding  []float64   `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	float32s := func(src []float64) []float32 {
		dst := make([]float32, len(src))
		for i, v := range src {
			dst[i] = float32(v)
		}
		return dst
	}

	switch {
	case len(parsed.Embeddings) > 0:
		out := make([][]float32, len(parsed.Embeddings))
		for i, emb := range parsed.Embeddings {
			out[i] = float32s(emb)
		}
		return out, nil
	case len(parsed.Embedding) > 0:
		return [][]float32{float32s(parsed.Embedding)}, nil
	default:
		return nil, errors.New("ollama embed returned no embeddings")
	}
}

// OllamaChatClient talks to Ollama's /api/chat endpoint.
type OllamaChatClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaChatClient constructs a chat client for Ollama.
func NewOllamaChatClient(baseURL, model string) *OllamaChatClient {
	if model == "" {
		model = DefaultOllamaChatModel
	}
	if baseURL == "" {
		baseURL = DefaultOllamaBaseURL
	}
	return &OllamaChatClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: 180 * time.Second}, // Increased to 3 minutes for LLM generation
	}
}

func (c *OllamaChatClient) Complete(ctx context.Context, systemPrompt, prompt string, temperature float32) (string, error) {
	if temperature == 0 {
		temperature = 0.2
	}
	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"stream":      false,
		"temperature": temperature,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Debug: log the model being used
	fmt.Printf("[DEBUG] Ollama chat request - Model: %s, URL: %s/api/chat\n", c.model, c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var parsed struct {
		Message *struct {
			Content string `json:"content"`
		} `json:"message"`
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	switch {
	case parsed.Message != nil:
		return strings.TrimSpace(parsed.Message.Content), nil
	case parsed.Response != "":
		return strings.TrimSpace(parsed.Response), nil
	default:
		return "", errors.New("ollama chat returned empty response")
	}
}
