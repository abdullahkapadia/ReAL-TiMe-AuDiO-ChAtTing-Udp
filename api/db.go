package api

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"go-udp-server/api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection pool
var DB *gorm.DB

// InitDB initializes the PostgreSQL database and auto-migrates the schema
func InitDB() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Construct PostgreSQL connection string securely from environment variables
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	fmt.Println("PostgreSQL database connection established!")

	// Auto-Migrate the schemas
	err = DB.AutoMigrate(&models.User{}, &models.Friendship{})
	if err != nil {
		log.Fatal("Failed to migrate database schemas:", err)
	}

	fmt.Println("Database schemas migrated successfully!")
}
