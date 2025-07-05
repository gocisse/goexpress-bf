package main

import (
	"fmt"
	"log"

	"goexpress-api/config"
	"goexpress-api/database"
	"goexpress-api/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Generate hash for admin password
	adminPassword := "goexpress123"
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	fmt.Printf("Password: %s\n", adminPassword)
	fmt.Printf("Hash: %s\n", hashedPassword)

	// Update or insert admin user
	_, err = db.Exec(`
		INSERT INTO users (name, email, password_hash, role) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) 
		DO UPDATE SET 
			password_hash = EXCLUDED.password_hash,
			updated_at = CURRENT_TIMESTAMP`,
		"GoExpress Admin", "admin@goexpress.com", hashedPassword, "admin")

	if err != nil {
		log.Fatal("Failed to create/update admin user:", err)
	}

	fmt.Println("✅ Admin user created/updated successfully!")
	fmt.Println("Email: admin@goexpress.com")
	fmt.Println("Password: goexpress123")
}

