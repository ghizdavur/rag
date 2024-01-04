package api

import (
	"cmd/main.go/pkg/repositories"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// HeaderLinks represents the structure of header links
type HeaderLinks struct {
	Login   string
	Contact string
	About   string
}

// SetupRoutes initializes and configures routes for the application
func SetupRoutes(app *fiber.App) {

	// Define statics - path to use - path in directories
	app.Static("/static", "../web/static/")
	app.Static("/assets", "../web/assets/")

	headerLinks := headerLinks()

	// Home Page
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// Login Page
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("register-login/login/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// Login Page
	app.Post("/login", func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		passwd := c.FormValue("passwd")
		var user repositories.User
		result := repositories.DB.Where("username = ?", username).First(&user)
		fmt.Println(user.Passwd)
		fmt.Println(passwd)
		fmt.Println(result)

		return c.Redirect("/success-test")
	})

	// Register Page
	app.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("register-login/register/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// Forgot password page Page
	app.Get("/forgot-password", func(c *fiber.Ctx) error {
		return c.Render("register-login/forgot-password/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// User Dashboard Page
	app.Get("/user-dashboard", func(c *fiber.Ctx) error {
		return c.Render("user-page/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// About Page
	app.Get("/about", func(c *fiber.Ctx) error {
		return c.Render("about/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})

	// Contact Page
	app.Get("/contact", func(c *fiber.Ctx) error {
		return c.Render("contact/index", fiber.Map{
			"HeaderLinksTab": headerLinks["HeaderLinksTab"],
		})
	})
}

func headerLinks() map[string][]HeaderLinks {
	return map[string][]HeaderLinks{
		"HeaderLinksTab": {
			{Login: "Login", Contact: "Contact", About: "About"},
		},
	}
}
