package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"cmd/main.go/pkg/api"
	"cmd/main.go/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	repositories.LoadEnvVariables()
	repositories.ConnectToDatabase()
}

func runMigrations(db *sql.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}

	fmt.Printf("Applied %d migrations!\n", n)
	return nil
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Access environment variables
	dbURL := os.Getenv("DB_URL")

	// Establish a connection to the PostgreSQL database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check the connection to the database
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the PostgreSQL database!")

	// Run Migrations
	if err := runMigrations(db); err != nil {
		log.Fatal(err)
	}

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
