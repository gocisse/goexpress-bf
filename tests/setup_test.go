package tests

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"goexpress-api/database"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *database.DB {
	// Use test PostgreSQL database
	testDatabaseURL := "postgres://goexpress:goexpress@localhost:5432/goexpress_test_db?sslmode=disable"

	db, err := database.New(testDatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up tables before each test
	_, err = db.Exec(`
		DROP TABLE IF EXISTS tracking_updates;
		DROP TABLE IF EXISTS shipments;
		DROP TABLE IF EXISTS zones;
		DROP TABLE IF EXISTS users;
	`)
	if err != nil {
		log.Printf("Warning: failed to clean up tables: %v", err)
	}

	// Run migrations
	migrationPath := filepath.Join("..", "supabase", "migrations", "20250704001632_weathered_block.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
