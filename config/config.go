package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	JWTSecret       string
	JWTRefreshSecret string
	Port            string
	Environment     string
	LogLevel        string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://goexpress:goexpress@localhost:5432/goexpress_db?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "goexpress-default-secret-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "goexpress-default-refresh-secret"),
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "production"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}