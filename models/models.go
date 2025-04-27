package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Movie struct {
	gorm.Model
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"` // in minutes
	Rating      float32 `json:"rating"`
	PosterURL   string  `json:"poster_url"`
	Shows       []Show  `json:"shows,omitempty"`
}

type Show struct {
	gorm.Model
	MovieID   uint      `json:"movie_id"`
	Movie     Movie     `json:"movie,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Screen    string    `json:"screen"`
	Price     float32   `json:"price"`
	Seats     []Seat    `json:"seats,omitempty"`
}

type Seat struct {
	gorm.Model
	ShowID    uint   `json:"show_id"`
	Show      Show   `json:"show,omitempty"`
	Row       string `json:"row"`
	Number    int    `json:"number"`
	Status    string `json:"status"` // available, booked, reserved
	BookingID *uint  `json:"booking_id,omitempty"`
}

type Booking struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	ShowID      uint    `json:"show_id"`
	Show        Show    `json:"show,omitempty"`
	Seats       []Seat  `json:"seats,omitempty"`
	TotalAmount float32 `json:"total_amount"`
	Status      string  `json:"status"` // confirmed, cancelled
	PaymentID   string  `json:"payment_id"`
}

type User struct {
	gorm.Model
	Email    string    `json:"email"`
	Password string    `json:"-"` // Password hash, not exposed in JSON
	Name     string    `json:"name"`
	Bookings []Booking `json:"bookings,omitempty"`
}
