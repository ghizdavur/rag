package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cmd/main.go/pkg/rag"

	"github.com/gofiber/fiber/v2"
	"github.com/jaytaylor/html2text"
)

// SetupRoutes initializes and configures routes for the RAG application
func SetupRoutes(app *fiber.App, ragService *rag.Service) {
	// Define statics - path to use - path in directories
	app.Static("/static", "./web/static/")
	app.Static("/assets", "./web/assets/")

	// RAG Home Page
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("rag/index", fiber.Map{})
	})

	// RAG API endpoint
	app.Post("/api/rag/query", func(c *fiber.Ctx) error {
		if ragService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "RAG service is not configured; run the ingestion workflow first.")
		}

		var request struct {
			Question string `json:"question"`
			TopK     int    `json:"topK"`
		}
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 180*time.Second) // Increased to 3 minutes for LLM generation
		defer cancel()

		answer, err := ragService.Answer(ctx, request.Question, rag.QueryOptions{TopK: request.TopK})
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}

		return c.JSON(answer)
	})

	// Add source endpoint (text or URL)
	app.Post("/api/rag/add-source", func(c *fiber.Ctx) error {
		if ragService == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "RAG service is not configured; run the ingestion workflow first.")
		}

		var request struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			URL     string `json:"url"`
		}
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 300*time.Second) // 5 minutes for adding source (includes embedding)
		defer cancel()

		// If URL is provided, fetch content
		content := strings.TrimSpace(request.Content)
		uri := strings.TrimSpace(request.URL)
		title := strings.TrimSpace(request.Title)

		// Debug logging
		fmt.Printf("Received request - Title: '%s', URL: '%s', Content length: %d\n", title, uri, len(content))

		// If both URL and content are empty, return error
		if uri == "" && content == "" {
			return fiber.NewError(fiber.StatusBadRequest, "either content or url must be provided")
		}

		// If URL is provided but no content, fetch from URL
		if uri != "" && content == "" {
			// Fetch from URL
			fetchedContent, err := fetchURLContent(ctx, uri)
			if err != nil {
				return fiber.NewError(fiber.StatusBadGateway, fmt.Sprintf("failed to fetch URL '%s': %v. Please try again or paste the content directly.", uri, err))
			}
			content = strings.TrimSpace(fetchedContent)
			if content == "" {
				return fiber.NewError(fiber.StatusBadGateway, fmt.Sprintf("URL '%s' returned empty content. Please paste the content directly instead.", uri))
			}
			if title == "" {
				title = "Source from " + uri
			}
		}

		// Final check - content should not be empty at this point
		if content == "" {
			return fiber.NewError(fiber.StatusBadRequest, "content cannot be empty. Please provide text content or a valid URL that returns content.")
		}

		if title == "" {
			title = "User Added Source"
		}
		if uri == "" {
			uri = "user-input://" + title
		}

		// Add source to store
		if err := ragService.AddSource(ctx, title, content, uri); err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}

		// Save updated store
		cfg := rag.LoadServiceConfigFromEnv()
		if err := ragService.SaveStore(cfg.IndexPath); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to save store: %v", err))
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Source added successfully",
		})
	})
}

// fetchURLContent fetches and converts content from a URL
func fetchURLContent(ctx context.Context, url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "RAG-Bot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "html") {
		// Convert HTML to text
		text, err := html2text.FromString(string(body), html2text.Options{PrettyTables: true})
		if err != nil {
			return "", err
		}
		return normalizeWhitespace(text), nil
	}

	// Assume plain text or markdown
	return normalizeWhitespace(string(body)), nil
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
