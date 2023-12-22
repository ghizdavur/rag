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
	// New engine(html)
	app := fiber.New(fiber.Config{
		Views:         html.New("../web/templates/", ".html"),
		CaseSensitive: false,
	})

	// Define statics - path to use - path in directories
	app.Static("/static", "../web/static/")
	app.Static("/assets", "../web/assets/")

	// Set up routes
	api.SetupRoutes(app)

	log.Fatal(app.Listen(":8000"))
}
