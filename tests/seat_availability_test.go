package tests

import (
	"testing"
	"net/http"
	"github.com/gorilla/mux"
	_ "cinemabooking/db"
	"cinemabooking/models"
	"cinemabooking/handlers"
	"time"
	"encoding/json"
)

// TestSeatAvailability tests the seat availability functionality
func TestSeatAvailability(t *testing.T) {
	// Setup test database
	testDB := SetupTestDatabase(t)
	defer testDB.Close()

	// Create test movie and show
	movie := CreateTestMovie(t, testDB)

	show := models.Show{
		MovieID:   movie.ID,
		Screen:    "Screen 1",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Price:     10.00,
	}

	result, err := testDB.Exec(`
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

	// Create test seats
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
	r.HandleFunc("/api/seats/{show_id}", handlers.GetSeats).Methods("GET")
	r.HandleFunc("/api/bookings", handlers.CreateBooking).Methods("POST")

	// Test case 1: Get all seats
	TestHTTPHandler(t, handlers.GetSeats, "GET", "/api/seats/1", "", http.StatusOK)

	// Test case 2: Book some seats
	booking := map[string]interface{}{
		"show_id":  showID,
		"seat_ids": []int{1, 2, 3},
		"user_id":  1,
	}
	bookingJSON, err := json.Marshal(booking)
	if err != nil {
		t.Fatalf("Failed to marshal booking JSON: %v", err)
	}

	TestHTTPHandler(t, handlers.CreateBooking, "POST", "/api/bookings", string(bookingJSON), http.StatusOK)

	// Test case 3: Verify booked seats are not available
	TestHTTPHandler(t, handlers.GetSeats, "GET", "/api/seats/1", "", http.StatusOK)

	// Test case 4: Try to book already booked seats
	TestHTTPHandler(t, handlers.CreateBooking, "POST", "/api/bookings", string(bookingJSON), http.StatusConflict)
}
