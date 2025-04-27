package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"cinemabooking/db"
	"cinemabooking/models"

	"github.com/gorilla/mux"
)

var (
	seatLocks = make(map[uint]*sync.Mutex)
	lockMutex sync.Mutex
)

func getSeatLock(seatID uint) *sync.Mutex {
	lockMutex.Lock()
	defer lockMutex.Unlock()

	if lock, exists := seatLocks[seatID]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	seatLocks[seatID] = lock
	return lock
}

func GetMovies(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	var movies []models.Movie
	if err := db.Find(&movies).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movies)
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := db.GetDB()
	var movie models.Movie
	if err := db.First(&movie, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

func GetShows(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	movieID := vars["id"]

	db := db.GetDB()
	var shows []models.Show
	if err := db.Where("movie_id = ?", movieID).Find(&shows).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(shows)
}

func GetSeats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	showID := vars["id"]

	db := db.GetDB()
	var seats []models.Seat
	if err := db.Where("show_id = ?", showID).Find(&seats).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(seats)
}

func CreateBooking(w http.ResponseWriter, r *http.Request) {
	var bookingRequest struct {
		ShowID  uint   `json:"show_id"`
		SeatIDs []uint `json:"seat_ids"`
		UserID  uint   `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&bookingRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	tx := db.Begin()

	// Check if seats are available and lock them
	for _, seatID := range bookingRequest.SeatIDs {
		lock := getSeatLock(seatID)
		lock.Lock()
		defer lock.Unlock()

		var seat models.Seat
		if err := tx.First(&seat, seatID).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Seat not found", http.StatusNotFound)
			return
		}

		if seat.Status != "available" {
			tx.Rollback()
			http.Error(w, "Seat already booked", http.StatusConflict)
			return
		}

		seat.Status = "reserved"
		if err := tx.Save(&seat).Error; err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Create booking
	booking := models.Booking{
		ShowID:      bookingRequest.ShowID,
		UserID:      bookingRequest.UserID,
		Status:      "confirmed",
		TotalAmount: 0, // Calculate based on show price
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update seats with booking ID
	if err := tx.Model(&models.Seat{}).Where("id IN (?)", bookingRequest.SeatIDs).
		Updates(map[string]interface{}{
			"status":     "booked",
			"booking_id": booking.ID,
		}).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(booking)
}

func GetBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db := db.GetDB()
	var booking models.Booking
	if err := db.Preload("Show").Preload("Seats").First(&booking, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(booking)
}
