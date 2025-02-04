package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
	templates   *template.Template

	voteCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vote_total",
			Help: "Total number of votes for each option.",
		},
		[]string{"option"},
	)
)

// func init() {
// 	// Register custom metrics
// 	prometheus.MustRegister(voteCount)
//     voteCount.With(prometheus.Labels{"option": "cat"}).Add(0)
//     voteCount.With(prometheus.Labels{"option": "dog"}).Add(0)
//  }

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
	// Test Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %s\n", err.Error())
	}

	// Parse templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Serve static files (CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/", votingPage)
	http.HandleFunc("/vote/cat", voteCat)
	http.HandleFunc("/vote/dog", voteDog)
	http.Handle("/metrics", promhttp.Handler())

	// Start server
	log.Println("Voting service is running on http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func votingPage(w http.ResponseWriter, r *http.Request) {
	// Get the results service URL from the environment variable
	resultsServiceURL := os.Getenv("RESULTS_SERVICE_URL")

	// Create data to pass to the template
	data := map[string]interface{}{
		"ResultsServiceURL": resultsServiceURL,
	}

	err := templates.ExecuteTemplate(w, "vote.html", data)
	if err != nil {
		http.Error(w, "Unable to load page", http.StatusInternalServerError)
	}
}

func voteCat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := redisClient.Incr(ctx, "cat_votes").Result()
		if err != nil {
			http.Error(w, "Unable to record vote", http.StatusInternalServerError)
			return
		}
		voteCount.With(prometheus.Labels{"option": "cat"}).Inc()
		notifyWorkerService()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func voteDog(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := redisClient.Incr(ctx, "dog_votes").Result()
		if err != nil {
			http.Error(w, "Unable to record vote", http.StatusInternalServerError)
			return
		}
		voteCount.With(prometheus.Labels{"option": "dog"}).Inc()
		notifyWorkerService()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func retrieveVotesFromRedis() (int, int, error) {
	catVotes, err := redisClient.Get(ctx, "cat_votes").Int()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}
	dogVotes, err := redisClient.Get(ctx, "dog_votes").Int()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}
	return catVotes, dogVotes, nil
}

func notifyWorkerService() {
	catVotes, dogVotes, err := retrieveVotesFromRedis()
	if err != nil {
		log.Println("Error retrieving votes from Redis:", err)
		return
	}

	data := map[string]int{
		"cat_votes": catVotes,
		"dog_votes": dogVotes,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error encoding data to JSON:", err)
		return
	}

	_, err = http.Post("http://worker-service:8084/sync", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error notifying worker service:", err)
	}
}