package createadmin
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/cody/lifesupport/internal/auth"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Get username and password from args
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run create_admin.go <username> <password>")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Insert user
	_, err = db.Exec(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET password_hash = $2",
		username, hash,
	)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ Admin user '%s' created/updated successfully\n", username)
	fmt.Printf("🔑 Password hash: %s\n", hash)
}
