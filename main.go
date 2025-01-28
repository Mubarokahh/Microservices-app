package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"  
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/lib/pq"
)
  
type VoteCounts struct {
	Id       int
	CatVotes int
	DogVotes int
}

var (
	templates  *template.Template
	postgresDB *sql.DB

	// Prometheus metrics
	resultsRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "results_requests_total",
		Help: "Total number of requests to the results endpoint.",
	})
	resultsRequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "results_request_duration_seconds",
		Help:    "Duration of requests to the results endpoint.",
		Buckets: prometheus.DefBuckets,
	})
)

// func init() {
// 	// Register metrics
// 	prometheus.MustRegister(resultsRequestCounter)
// 	prometheus.MustRegister(resultsRequestDuration)
// }

func main() {
	// Initialize PostgreSQL connection and create database/table
	initPostgres()

	// Parse templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Serve static files (CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/results", showResults)

	http.Handle("/metrics", promhttp.Handler())

	// Start server
	log.Println("Results service is running on http://localhost:8085")
	serverErr := http.ListenAndServe(":8085", nil)
	if serverErr != nil {
		log.Fatalf("Could not start server: %s\n", serverErr.Error())
	}
}

func initPostgres() {
	var err error
	// envErr := godotenv.Load(".env")
	// if envErr != nil {
	// 	log.Fatalf("Error loading .env file")
	// }

	//Retrieve environment variables
	// POSTGRES_HOST := os.Getenv("POSTGRES_HOST")
	// POSTGRES_USER := os.Getenv("POSTGRES_USER")
	// POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	// POSTGRES_DB_NEW := os.Getenv("POSTGRES_DB_NEW")
	// POSTGRES_DB := os.Getenv("POSTGRES_DB")

	//Declaring Variables
	POSTGRES_HOST := "postgres"
	POSTGRES_USER := "postgres"
	POSTGRES_PASSWORD := "Barakat1243"
	POSTGRES_DB_NEW := "votingdb"
	POSTGRES_DB := "postgres"

	// Step 1: Connect to the default "postgres" database to check if "votingdb" exists
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", POSTGRES_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)

	postgresDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %s\n", err.Error())
	}
	defer postgresDB.Close()

	// Step 2: Check if the "votingdb" database exists
	var exists bool
	err = postgresDB.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'votingdb')").Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking for database existence: %s\n", err.Error())
	}

	// Step 3: Create the "votingdb" database if it doesn't exist
	if !exists {
		_, err = postgresDB.Exec("CREATE DATABASE votingdb")
		if err != nil {
			log.Fatalf("Error creating database: %s\n", err.Error())
		}
		log.Println("Database 'votingdb' created successfully.")
	} else {
		log.Println("Database 'votingdb' already exists.")
	}

	// Step 4: Connect to the "votingdb" database
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

func showResults(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
	resultsRequestCounter.Inc() // Count request
	defer resultsRequestDuration.Observe(time.Since(start).Seconds()) // Observe duration

    // Get the voting page URL from the environment
    votingServiceURL := os.Getenv("VOTING_SERVICE_URL")
    if votingServiceURL == "" {
        votingServiceURL = "http://localhost:8083"  // Fallback URL in case env is not set
    }

    // Retrieve the votes from PostgreSQL
    row := postgresDB.QueryRow("SELECT id, cat_votes, dog_votes FROM votes")
    voteCounts := VoteCounts{}
    err := row.Scan(&voteCounts.Id, &voteCounts.CatVotes, &voteCounts.DogVotes)
    if err != nil {
        http.Error(w, "Unable to retrieve votes", http.StatusInternalServerError)
        return
    }

    // Combine voteCounts and votingServiceURL into a struct
    data := struct {
        VoteCounts
        VotingServiceURL string
    }{
        VoteCounts:       voteCounts,
        VotingServiceURL: votingServiceURL,
    }

    // Render the template with the voteCounts and URL
    err = templates.ExecuteTemplate(w, "results.html", data)
    if err != nil {
        http.Error(w, "Unable to load results", http.StatusInternalServerError)
    }
}
  
  