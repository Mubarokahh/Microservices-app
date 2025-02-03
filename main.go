package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	postgresDB  *sql.DB
	redisClient *redis.Client
	ctx         = context.Background()

	syncCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vote_sync_attempts_total",
		Help: "Total number of attempts to sync votes from Redis to PostgreSQL.",
	})
	syncErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vote_sync_errors_total",
		Help: "Total number of sync errors when transferring votes to PostgreSQL.",
	})
)

// func init() {
// 	// Register Prometheus metrics
// 	prometheus.MustRegister(syncCounter)
// 	prometheus.MustRegister(syncErrorCounter)
// }

func main() {
	redisErr := godotenv.Load()
	if redisErr != nil {
		log.Fatalf("Error loading .env file")
	}

	//REDIS_ADDR := os.Getenv("REDIS_ADDR")
	REDIS_ADDR := "redis:6379"

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: REDIS_ADDR,
	})
	// Initialize PostgreSQL connection and create database/table
	initPostgres()

	http.HandleFunc("/sync", syncVotesHandler)
  
	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Worker service is running on http://localhost:8084")
	serverErr := http.ListenAndServe(":8084", nil)
	if serverErr != nil {
		log.Fatalf("Could not start server: %s\n", serverErr.Error())
	}
}

func initPostgres() {
	log.Println("Connecting to PostgreSQL...")
	var err error
	envErr := godotenv.Load(".env")
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}

	//Retrieve environment variables
	POSTGRES_HOST := os.Getenv("POSTGRES_HOST")
	POSTGRES_USER := os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_DB_NEW := os.Getenv("POSTGRES_DB_NEW")
	POSTGRES_DB := os.Getenv("POSTGRES_DB")

	///POSTGRES_HOST := "postgres"
	//POSTGRES_USER := "postgres"
	//POSTGRES_PASSWORD := "Barakat1243"
	//POSTGRES_DB_NEW := "votingdb"
	//POSTGRES_DB := "postgres"

	// Step 1: Connect to the default "postgres" database

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", POSTGRES_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)

	postgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %s\n", err.Error())
	}
	log.Println("Connected to PostgreSQL successfully.")
	defer postgresDB.Close()

	// Step 2: Check if the "votingdb" database exists
	var exists bool
	err = postgresDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'votingdb')").Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking for database existence: %s\n", err.Error())
	}

	// Step 3: Create the database if it doesn't exist
	if !exists {
		_, err = postgresDB.Exec("CREATE DATABASE votingdb")
		if err != nil {
			log.Fatalf("Error creating database: %s\n", err.Error())
		}
		log.Println("Database 'votingdb' created successfully.")
	} else {
		log.Println("Database 'votingdb' already exists.")
	}

	// Step 4: Now connect to the `votingdb` database
	connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", POSTGRES_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB_NEW)
	postgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to votingdb: %s\n", err.Error())
	}

	// Step 5: Create the votes table if it doesn't exist
	_, err = postgresDB.Exec(`
        CREATE TABLE IF NOT EXISTS votes (
            id SERIAL PRIMARY KEY,
            cat_votes INTEGER DEFAULT 0,
            dog_votes INTEGER DEFAULT 0
        )
    `)
	if err != nil {
		log.Fatalf("Error creating votes table: %s\n", err.Error())
	}

	log.Println("PostgreSQL database and table initialized successfully.")
}

func RetrieveVotesFromRedis() (catVotes, dogVotes int, err error) {
	log.Println("Connecting to Redis to retrieve votes...")
	catVotes, err = redisClient.Get(ctx, "cat_votes").Int()
	if err != nil && err != redis.Nil {
		log.Printf("Error retrieving cat votes: %s\n", err.Error())
		return 0, 0, err
	}
	log.Printf("Retrieved cat votes: %d\n", catVotes)

	dogVotes, err = redisClient.Get(ctx, "dog_votes").Int()
	if err != nil && err != redis.Nil {
		log.Printf("Error retrieving dog votes: %s\n", err.Error())
		return 0, 0, err
	}
	log.Printf("Retrieved dog votes: %d\n", dogVotes)
	return catVotes, dogVotes, nil
}

func syncVotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		syncCounter.Inc() // Increment sync attempt count
		log.Println("Received request to sync votes from Redis to PostgreSQL...")
		catVotes, dogVotes, err := RetrieveVotesFromRedis()
		if err != nil {
			log.Println("Received request to sync votes from Redis to PostgreSQL...")
			syncErrorCounter.Inc() // Increment error counter on failure
			http.Error(w, "Unable to retrieve votes", http.StatusInternalServerError)
			return
		}

		log.Printf("Syncing cat votes: %d, dog votes: %d to PostgreSQL...\n", catVotes, dogVotes)

		_, err = postgresDB.Exec(`
            INSERT INTO votes (id, cat_votes, dog_votes)
            VALUES (1, $1, $2)
            ON CONFLICT (id) DO UPDATE 
            SET cat_votes = EXCLUDED.cat_votes, dog_votes = EXCLUDED.dog_votes`,
			catVotes, dogVotes)

		if err != nil {
			log.Printf("Error syncing votes to PostgreSQL: %s\n", err.Error())
			http.Error(w, "Unable to sync votes", http.StatusInternalServerError)
			return
		}
		log.Println("Votes synced to PostgreSQL successfully.")
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}