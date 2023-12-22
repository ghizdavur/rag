package main

import (
	"log"

	"cmd/main.go/pkg/api"
	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
)

func init() {
	repositories.LoadEnvVariables()
	repositories.ConnectToDatabase()
	repositories.RunMigrations(repositories.DB)
}

func main() {
	app := fiber.New(fiber.Config{
		Views:         html.New("../web/templates/", ".html"),
		CaseSensitive: false,
	})

	api.SetupRoutes(app)

	log.Fatal(app.Listen(":8000"))
}
