package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var dbConn *sql.DB

func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName
	dbConn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func seedDatabase() {
	// Sample movies
	movies := []struct {
		Title       string
		Description string
		Duration    int
		Rating      string
		PosterURL   string
	}{
		{
			Title:       "The Dark Knight",
			Description: "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.",
			Duration:    152,
			Rating:      "9.0",
			PosterURL:   "https://m.media-amazon.com/images/M/MV5BMTMxNTMwODM0NF5BMl5BanBnXkFtZTcwODAyMTk2Mw@@._V1_.jpg",
		},
		{
			Title:       "Inception",
			Description: "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
			Duration:    148,
			Rating:      "8.8",
			PosterURL:   "https://m.media-amazon.com/images/M/MV5BMjAxMzY3NjcxNF5BMl5BanBnXkFtZTcwNTI5OTM0Mw@@._V1_.jpg",
		},
		{
			Title:       "The Shawshank Redemption",
			Description: "Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
			Duration:    142,
			Rating:      "9.3",
			PosterURL:   "https://m.media-amazon.com/images/M/MV5BNDE3ODcxYzMtY2YzZC00NmNlLWJiNDMtZDViZWM2MzIxZDYwXkEyXkFqcGdeQXVyNjAwNDUxODI@._V1_.jpg",
		},
	}

	// Create movies
	for _, movie := range movies {
		result, err := dbConn.Exec(`
			INSERT INTO movies (title, description, duration, rating, poster_url)
			VALUES (?, ?, ?, ?, ?)
		`, movie.Title, movie.Description, movie.Duration, movie.Rating, movie.PosterURL)
		if err != nil {
			log.Printf("Error creating movie: %v", err)
			continue
		}

		movieID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting movie ID: %v", err)
			continue
		}

		// Create shows for each movie
		shows := []struct {
			StartTime time.Time
			EndTime   time.Time
			Screen    int
			Price     float64
		}{
			{
				StartTime: time.Now().Add(24 * time.Hour),
				EndTime:   time.Now().Add(26 * time.Hour),
				Screen:    1,
				Price:     12.99,
			},
			{
				StartTime: time.Now().Add(48 * time.Hour),
				EndTime:   time.Now().Add(50 * time.Hour),
				Screen:    2,
				Price:     14.99,
			},
		}

		for _, show := range shows {
			result, err := dbConn.Exec(`
				INSERT INTO shows (movie_id, screen, start_time, end_time, price)
				VALUES (?, ?, ?, ?, ?)
			`, movieID, show.Screen, show.StartTime, show.EndTime, show.Price)
			if err != nil {
				log.Printf("Error creating show: %v", err)
				continue
			}

			showID, err := result.LastInsertId()
			if err != nil {
				log.Printf("Error getting show ID: %v", err)
				continue
			}

			// Create 50 seats for each show
			rows := []string{"A", "B", "C", "D", "E"}
			seatsPerRow := 10

			for _, row := range rows {
				for seatNumber := 1; seatNumber <= seatsPerRow; seatNumber++ {
					_, err := dbConn.Exec(`
						INSERT INTO seats (show_id, row, seat_number, status)
						VALUES (?, ?, ?, ?)
					`, showID, row, seatNumber, "available")
					if err != nil {
						log.Printf("Error creating seat: %v", err)
					}
				}
			}
		}
	}

	log.Println("Database seeded successfully!")
}

func main() {
	initDB()
	defer dbConn.Close()

	seedDatabase()
}
