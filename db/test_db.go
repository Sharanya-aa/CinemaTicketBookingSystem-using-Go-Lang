package db

import (
	"database/sql"
	"testing"
	"github.com/go-sql-driver/mysql"
	"os"
	"log"
)

// TestDBConfig contains test database configuration
var TestDBConfig = mysql.Config{
	User:   os.Getenv("DB_USER"),
	Passwd: os.Getenv("DB_PASSWORD"),
	Net:    "tcp",
	Addr:   os.Getenv("DB_HOST"),
	DBName: "test_cinema_db",
	ParseTime: true,
}

// GetTestDB returns a test database connection
func GetTestDB(t *testing.T) *sql.DB {
	// First connect without database name
	baseConfig := TestDBConfig
	baseConfig.DBName = ""
	dsn := baseConfig.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}

	// Create test database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS test_cinema_db")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Use the test database
	_, err = db.Exec("USE test_cinema_db")
	if err != nil {
		t.Fatalf("Failed to use test database: %v", err)
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS movies (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			duration INT NOT NULL,
			rating FLOAT NOT NULL,
			poster_url VARCHAR(255)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create movies table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS shows (
			id INT AUTO_INCREMENT PRIMARY KEY,
			movie_id INT NOT NULL,
			screen VARCHAR(50) NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			FOREIGN KEY (movie_id) REFERENCES movies(id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create shows table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS seats (
			id INT AUTO_INCREMENT PRIMARY KEY,
			show_id INT NOT NULL,
			row_name VARCHAR(10) NOT NULL,
			seat_number INT NOT NULL,
			status VARCHAR(20) NOT NULL,
			booking_id INT,
			FOREIGN KEY (show_id) REFERENCES shows(id),
			FOREIGN KEY (booking_id) REFERENCES bookings(id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create seats table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookings (
			id INT AUTO_INCREMENT PRIMARY KEY,
			show_id INT NOT NULL,
			user_id INT NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) NOT NULL,
			FOREIGN KEY (show_id) REFERENCES shows(id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create bookings table: %v", err)
	}

	// Clean up existing data
	_, err = db.Exec("DELETE FROM bookings")
	if err != nil {
		t.Fatalf("Failed to clean up bookings: %v", err)
	}

	_, err = db.Exec("DELETE FROM seats")
	if err != nil {
		t.Fatalf("Failed to clean up seats: %v", err)
	}

	_, err = db.Exec("DELETE FROM shows")
	if err != nil {
		t.Fatalf("Failed to clean up shows: %v", err)
	}

	_, err = db.Exec("DELETE FROM movies")
	if err != nil {
		t.Fatalf("Failed to clean up movies: %v", err)
	}

	return db
}
