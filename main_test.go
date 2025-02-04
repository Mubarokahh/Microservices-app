// voting_service_test.go
package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_votingPage(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new recorder (simulating HTTP response)
	rr := httptest.NewRecorder()

	// Mock templates
	templates = template.Must(template.New("vote.html").Parse("Vote Page"))

	// Call the handler function
	http.HandlerFunc(votingPage).ServeHTTP(rr, req)

	// Check if the status code is what we expect
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}  
}    