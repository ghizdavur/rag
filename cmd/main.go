package main

import (
	"log"

	"cmd/main.go/pkg/api"
	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func init() {
	repositories.LoadEnvVariables()
	repositories.ConnectToDatabase()
}

func main() {
	// New engine(html)
	app := fiber.New(fiber.Config{
		Views:         html.New("../web/templates/", ".html"),
		CaseSensitive: false,
	})

	// Define statics - path to use - path in directories
	app.Static("/static", "../web/static/")

	// Set up routes
	api.SetupRoutes(app)

	log.Fatal(app.Listen(":8000"))
}
