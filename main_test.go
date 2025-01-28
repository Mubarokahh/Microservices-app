package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Test initPostgres function
// func Test_initPostgres(t *testing.T) {
// 	// Mock environment variables
// 	// os.Setenv("POSTGRES_HOST", "localhost")
// 	// os.Setenv("POSTGRES_USER", "postgres")
// 	// os.Setenv("POSTGRES_PASSWORD", "Barakat1243")
// 	// os.Setenv("POSTGRES_DB_NEW", "votingdb")
// 	// os.Setenv("POSTGRES_DB", "postgres")

// 	// POSTGRES_HOST := "localhost"
// 	// POSTGRES_USER := "postgres"
// 	// POSTGRES_PASSWORD := "Barakat1243"
// 	// POSTGRES_DB_NEW := "votingdb"
// 	// POSTGRES_DB := "postgres"

// 	// Mock PostgreSQL connection
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer db.Close()

// 	// Set the mock database connection
// 	postgresDB = db

// 	// Mock the check for "votingdb" existence
// 	mock.ExpectQuery(`SELECT EXISTS\(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'votingdb'\)`).WillReturnRows(
// 		sqlmock.NewRows([]string{"exists"}).AddRow(false)) // Assume database does not exist

// 	// Mock the creation of the database
// 	mock.ExpectExec(`CREATE DATABASE votingdb`).WillReturnResult(sqlmock.NewResult(0, 0)) // `LastInsertId` and `RowsAffected` values are not used for this command

// 	// Mock the connection to "votingdb"
// 	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS votes`).WillReturnResult(sqlmock.NewResult(0, 0)) // `LastInsertId` and `RowsAffected` values are not used for this command

// 	// Call the function that initializes the DB
// 	initPostgres()
// 	// Ensure that all expectations were met
// 	// if err := mock.ExpectationsWereMet(); err != nil {
// 	// 	t.Errorf("there were unmet expectations: %v", err)
// 	// }
// }

// Test showResults handler
func Test_showResults(t *testing.T) {
	// Step 1: Initialize the mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	// Set the global postgresDB to our mock DB
	postgresDB = db

	// Step 2: Expect the query that retrieves vote counts from the votes table
	// Simulate the expected row (id=1, cat_votes=50, dog_votes=40)
	mock.ExpectQuery("SELECT id, cat_votes, dog_votes FROM votes").WillReturnRows(
		sqlmock.NewRows([]string{"id", "cat_votes", "dog_votes"}).AddRow(1, 50, 40),
	)

	// Step 3: Create a mock HTTP request to simulate a request to the results endpoint
	req, err := http.NewRequest("GET", "/results", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Step 4: Create a response recorder to capture the handler's response
	rr := httptest.NewRecorder()

	// Step 5: Initialize the template (mock)
	templates = template.Must(template.New("results.html").Parse(`{{.CatVotes}} Cats, {{.DogVotes}} Dogs`))

	// Step 6: Call the showResults handler
	handler := http.HandlerFunc(showResults)
	handler.ServeHTTP(rr, req)

	// Step 7: Check if the status code is 200 (OK)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Step 8: Verify that the mock query was actually executed
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unmet expectations: %v", err)
	}
}

// Test when PostgreSQL query fails in showResults
func Test_showResultsQueryFailure(t *testing.T) {
	// Mock PostgreSQL connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Mock query failure
	mock.ExpectQuery("SELECT id, cat_votes, dog_votes FROM votes").WillReturnError(sql.ErrNoRows)

	// Set the mock DB as the actual database
	postgresDB = db

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/results", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()

	// Call the showResults handler
	handler := http.HandlerFunc(showResults)
	handler.ServeHTTP(rr, req)

	// Check if the status code is what we expect (500 Internal Server Error)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}  