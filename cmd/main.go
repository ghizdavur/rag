package main

// Main entry point for the web app
import (
	"log"

	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func init() {
	repositories.LoadEnvVariables()
	repositories.ConnectToDatabase()
}

type HeaderLinks struct {
	Login   string
	Contact string
	About   string
}

func main() {
	// New engine(html)
	app := fiber.New(fiber.Config{
		Views: html.New("../web/templates/", ".html"),
	})

	// Define statics - path to use - path in directories
	app.Static("/static", "../web/static/")

	// Routes
	headerLinks := headerLinks()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("login/index", fiber.Map{
			"HeaderLinks": headerLinks["HeaderLinksTab"],
		})
	})

	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about/index", fiber.Map{
			"HeaderLinks": headerLinks["HeaderLinksTab"],
		})
	})

	app.Get("/contact", func(c *fiber.Ctx) error {
		return c.Render("contact/index", fiber.Map{
			"HeaderLinks": headerLinks["HeaderLinksTab"],
		})
	})

	log.Fatal(app.Listen(":8000"))
}

func headerLinks() map[string][]HeaderLinks {
	return map[string][]HeaderLinks{
		"HeaderLinksTab": {
			{Login: "Login", Contact: "Contact", About: "About"},
		},
	}
}
