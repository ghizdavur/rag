package main

import (
	"context"
	"log"
	"path/filepath"

	"cmd/main.go/pkg/api"
	"cmd/main.go/pkg/config"
	"cmd/main.go/pkg/rag"
	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func init() {
	config.LoadEnvVariables()
	// Connect to Supabase (optional - will log error if connection fails but won't stop the app)
	if err := repositories.ConnectToDatabase(); err != nil {
		log.Printf("Database connection failed (optional): %v", err)
		log.Printf("RAG will work without database")
	} else {
		log.Printf("✅ Connected to Supabase database")
	}
}

func main() {
	templatePath, err := filepath.Abs("./web/templates/")
	if err != nil {
		log.Fatalf("Error resolving template path: %v", err)
	}

	engine := html.New(templatePath, ".html")
	engine.Reload(true) // Enable auto-reload for development

	app := fiber.New(fiber.Config{
		Views:         engine,
		CaseSensitive: false,
	})

	ctx := context.Background()
	ragService, err := rag.NewServiceFromEnv(ctx)
	if err != nil {
		log.Printf("RAG service disabled: %v", err)
	}

	api.SetupRoutes(app, ragService)

	log.Fatal(app.Listen(":8000"))
}
