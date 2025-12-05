package utils

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadTestEnv() {
	// Load .env file if present
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found, using OS environment variables")
	}
}
