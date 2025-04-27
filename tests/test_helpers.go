package tests

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"cinemabooking/db"
	"cinemabooking/models"
	"strings"
	"database/sql"
)

// TestHTTPHandler tests an HTTP handler with a test request
func TestHTTPHandler(t *testing.T, handler func(http.ResponseWriter, *http.Request), method, path, body string, expectedStatus int) {
	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc(path, handler).Methods(method)

	r.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, rec.Code)
	}
}

// SetupTestDatabase sets up a clean test database
func SetupTestDatabase(t *testing.T) *sql.DB {
	testDB := db.GetTestDB(t)
	if testDB == nil {
		t.Fatal("Failed to get test database connection")
	}

	// Clean up existing data
	_, err := testDB.Exec("DELETE FROM bookings")
	if err != nil {
		t.Fatal(err)
	}

	_, err = testDB.Exec("DELETE FROM seats")
	if err != nil {
		t.Fatal(err)
	}

	_, err = testDB.Exec("DELETE FROM shows")
	if err != nil {
		t.Fatal(err)
	}

	_, err = testDB.Exec("DELETE FROM movies")
	if err != nil {
		t.Fatal(err)
	}

	return testDB
}

// CreateTestMovie creates a test movie in the database
func CreateTestMovie(t *testing.T, db *sql.DB) models.Movie {
	movie := models.Movie{
		Title:       "Test Movie",
		Description: "Test description",
		Duration:    120,
		Rating:      8.5,
		PosterURL:   "",
	}

	result, err := db.Exec(`
		INSERT INTO movies (title, description, duration, rating, poster_url)
		VALUES (?, ?, ?, ?, ?)
	`, movie.Title, movie.Description, movie.Duration, movie.Rating, movie.PosterURL)

	if err != nil {
		t.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	movie.ID = uint(id)
	return movie
}
