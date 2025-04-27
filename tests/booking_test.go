package tests

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"cinemabooking/db"
	"cinemabooking/models"
	"cinemabooking/handlers"
	"time"
	"strconv"
	"strings"
)

// TestCreateBooking tests the seat booking functionality
func TestCreateBooking(t *testing.T) {
	// Setup test database
	testDB := db.GetTestDB(t)
	defer testDB.Close()

	// Create test data
	movie := models.Movie{
		Title:       "Test Movie",
		Description: "Test description",
		Duration:    120,
		Rating:      8.5,
		PosterURL:   "",
	}

	// Insert test movie
	result, err := testDB.Exec(`
		INSERT INTO movies (title, description, duration, rating, poster_url)
		VALUES (?, ?, ?, ?, ?)
	`, movie.Title, movie.Description, movie.Duration, movie.Rating, movie.PosterURL)
	if err != nil {
		t.Fatalf("Failed to insert test movie: %v", err)
	}

	movieID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get movie ID: %v", err)
	}

	// Insert test show
	show := models.Show{
		MovieID:   uint(movieID),
		Screen:    "Screen 1",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Price:     10.00,
	}

	result, err = testDB.Exec(`
		INSERT INTO shows (movie_id, screen, start_time, end_time, price)
		VALUES (?, ?, ?, ?, ?)
	`, show.MovieID, show.Screen, show.StartTime.Format("2006-01-02 15:04:05"), show.EndTime.Format("2006-01-02 15:04:05"), show.Price)
	if err != nil {
		t.Fatalf("Failed to insert test show: %v", err)
	}

	showID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get show ID: %v", err)
	}

	// Insert test seats
	for i := 1; i <= 10; i++ {
		_, err := testDB.Exec(`
			INSERT INTO seats (show_id, row_name, seat_number, status)
			VALUES (?, ?, ?, ?)
		`, showID, "A", i, "available")
		if err != nil {
			t.Fatalf("Failed to insert test seat %d: %v", i, err)
		}
	}

	// Create test router
	r := mux.NewRouter()
	r.HandleFunc("/api/bookings", handlers.CreateBooking).Methods("POST")

	// Test case 1: Successful booking
	testBooking := `{
		"show_id": ` + strconv.FormatInt(showID, 10) + `,
		"seat_ids": [1, 2, 3],
		"user_id": 1
	}`

	req, _ := http.NewRequest("POST", "/api/bookings", strings.NewReader(testBooking))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Test case 2: Booking already booked seats
	req, _ = http.NewRequest("POST", "/api/bookings", strings.NewReader(testBooking))
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rec.Code)
	}

	// Test case 3: Invalid show ID
	invalidBooking := `{
		"show_id": 999,
		"seat_ids": [1, 2, 3],
		"user_id": 1
	}`

	req, _ = http.NewRequest("POST", "/api/bookings", strings.NewReader(invalidBooking))
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}
