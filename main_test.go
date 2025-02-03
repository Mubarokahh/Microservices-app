package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"

	//"os"
	"testing"

	"github.com/alicebob/miniredis/v2" // mock Redis server
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Mock PostgreSQL setup
func setupMockPostgres(t *testing.T) *sql.DB {
	// envErr := godotenv.Load(".env")
	// if envErr != nil {
	// 	log.Fatalf("Error loading .env file")
	// }

	// POSTGRES_USER := os.Getenv("POSTGRES_USER")
	// POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	// POSTGRES_DB := os.Getenv("POSTGRES_DB")

	POSTGRES_USER := "postgres"
	POSTGRES_PASSWORD := "Barakat1243"
	POSTGRES_DB := "postgres"

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)

	postgresDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to mock PostgreSQL: %v", err)
	}

	// Create the `votes` table if it doesn't exist
	_, err = postgresDB.Exec(`
		CREATE TABLE IF NOT EXISTS votes (
			id SERIAL PRIMARY KEY,
			cat_votes INTEGER DEFAULT 0,
			dog_votes INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create votes table: %v", err)
	}

	return postgresDB
}

func Test_syncVotesHandler(t *testing.T) {
	// Setup mock Redis server
	mockRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start mock Redis: %v", err)
	}
	defer mockRedis.Close()

	// Initialize Redis client pointing to mock Redis server
	redisClient = redis.NewClient(&redis.Options{
		Addr: mockRedis.Addr(),
	})
	defer redisClient.Close()

	// Setup mock PostgreSQL connection
	postgresDB = setupMockPostgres(t)
	defer postgresDB.Close()

	// Set initial data in the mock Redis server
	mockRedis.Set("cat_votes", "5")
	mockRedis.Set("dog_votes", "3")

	// Create a new request and response recorder
	req, err := http.NewRequest("POST", "/sync", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Call the handler function
	http.HandlerFunc(syncVotesHandler).ServeHTTP(rr, req)

	// Check if the status code is what we expect
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify that votes were synced correctly in PostgreSQL
	var catVotes, dogVotes int
	err = postgresDB.QueryRow("SELECT cat_votes, dog_votes FROM votes WHERE id = 1").Scan(&catVotes, &dogVotes)
	if err != nil {
		t.Fatalf("Failed to retrieve votes from PostgreSQL: %v", err)
	}

	if catVotes != 5 {
		t.Errorf("Expected 5 cat votes, got %d", catVotes)
	}
	if dogVotes != 3 {
		t.Errorf("Expected 3 dog votes, got %d", dogVotes)
	}
}  