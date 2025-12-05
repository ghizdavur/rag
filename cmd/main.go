package main

import (
	"context"
	"log"

	"cmd/main.go/cmd/migrations"
	"cmd/main.go/pkg/api"
	"cmd/main.go/pkg/config"
	"cmd/main.go/pkg/rag"
	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
)

func init() {
	config.LoadEnvVariables()
	repositories.ConnectToDatabase()
	migrations.RunMigrations(repositories.DB)
}

func main() {
	app := fiber.New(fiber.Config{
		Views:         html.New("../web/templates/", ".html"),
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
