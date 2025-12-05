package rag

import (
	"context"
	"errors"
	"fmt"
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
		model = DefaultEmbeddingModel
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
		model = DefaultChatModel
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
