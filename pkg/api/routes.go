package api

import (
	"context"
	"time"

	"cmd/main.go/pkg/rag"

	"github.com/gofiber/fiber/v2"
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
		ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()

		answer, err := ragService.Answer(ctx, request.Question, rag.QueryOptions{TopK: request.TopK})
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}

		return c.JSON(answer)
	})
}
