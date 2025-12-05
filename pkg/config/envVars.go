package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v. Using environment variables from system.", err)
	}
}
