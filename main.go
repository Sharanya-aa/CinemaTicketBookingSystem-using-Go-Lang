package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Database connection
var dbConn *sql.DB

// Initialize database connection
func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName + "?parseTime=true"
	dbConn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to database")

	// Check if tables already exist
	var tableCount int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name IN ('movies', 'shows', 'seats', 'bookings')", dbName).Scan(&tableCount)
	if err != nil {
		log.Fatal("Error checking if tables exist:", err)
	}

	// Only create tables if they don't exist
	if tableCount < 4 {
		log.Println("Creating database tables...")

		// Create tables without foreign keys first
		_, err = dbConn.Exec(`
			CREATE TABLE IF NOT EXISTS movies (
				id INT AUTO_INCREMENT PRIMARY KEY,
				title VARCHAR(255) NOT NULL,
				description TEXT,
				duration INT NOT NULL,
				rating VARCHAR(10),
				poster_url TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatal("Error creating movies table:", err)
		}

		_, err = dbConn.Exec(`
			CREATE TABLE IF NOT EXISTS shows (
				id INT AUTO_INCREMENT PRIMARY KEY,
				movie_id INT NOT NULL,
				screen VARCHAR(50) NOT NULL,
				start_time DATETIME NOT NULL,
				end_time DATETIME NOT NULL,
				price DECIMAL(10,2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatal("Error creating shows table:", err)
		}

		// Create bookings table
		_, err = dbConn.Exec(`
			CREATE TABLE IF NOT EXISTS bookings (
				id INT AUTO_INCREMENT PRIMARY KEY,
				show_id INT NOT NULL,
				user_name VARCHAR(255),
				user_email VARCHAR(255),
				total_amount DECIMAL(10,2) NOT NULL,
				booking_time DATETIME NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatal("Error creating bookings table:", err)
		}

		// Create the seats table
		_, err = dbConn.Exec(`
			CREATE TABLE IF NOT EXISTS seats (
				id INT AUTO_INCREMENT PRIMARY KEY,
				show_id INT NOT NULL,
				row_name VARCHAR(1) NOT NULL,
				seat_number INT NOT NULL,
				status VARCHAR(20) NOT NULL DEFAULT 'available',
				booking_id INT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			log.Fatal("Error creating seats table:", err)
		}

		// Add foreign key constraints
		_, err = dbConn.Exec(`
			ALTER TABLE shows
			ADD CONSTRAINT fk_shows_movie_id
			FOREIGN KEY (movie_id) REFERENCES movies(id)
		`)
		if err != nil {
			log.Printf("Error adding foreign key to shows table (may already exist): %v", err)
		}

		_, err = dbConn.Exec(`
			ALTER TABLE seats
			ADD CONSTRAINT fk_seats_show_id
			FOREIGN KEY (show_id) REFERENCES shows(id)
		`)
		if err != nil {
			log.Printf("Error adding foreign key to seats table (may already exist): %v", err)
		}

		_, err = dbConn.Exec(`
			ALTER TABLE bookings
			ADD CONSTRAINT fk_bookings_show_id
			FOREIGN KEY (show_id) REFERENCES shows(id)
		`)
		if err != nil {
			log.Printf("Error adding foreign key to bookings table (may already exist): %v", err)
		}

		_, err = dbConn.Exec(`
			ALTER TABLE seats
			ADD CONSTRAINT fk_seats_booking_id
			FOREIGN KEY (booking_id) REFERENCES bookings(id)
		`)
		if err != nil {
			log.Printf("Error adding foreign key from seats to bookings (may already exist): %v", err)
		}
	} else {
		log.Println("Database tables already exist")
	}

	log.Println("Database tables created successfully")
}

type Movie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Rating      string `json:"rating"`
	PosterURL   string `json:"poster_url"`
}

type Show struct {
	ID        int     `json:"id"`
	MovieID   int     `json:"movie_id"`
	Screen    string  `json:"screen"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Price     float64 `json:"price"`
	Duration  int     `json:"duration"`
}

type Seat struct {
	ID         int    `json:"id"`
	ShowID     int    `json:"show_id"`
	Row        string `json:"row" db:"row_name"`
	SeatNumber int    `json:"seat_number"`
	Status     string `json:"status"`
}

// Add booking struct
type Booking struct {
	ID          int     `json:"id"`
	ShowID      int     `json:"show_id"`
	UserName    string  `json:"user_name"`
	UserEmail   string  `json:"user_email"`
	SeatIDs     []int   `json:"seat_ids"`
	TotalAmount float64 `json:"total_amount"`
	BookingTime string  `json:"booking_time"`
	Status      string  `json:"status"`
}

func seedDB() {
	log.Println("Starting database seeding...")

	// Check if tables exist and have data
	var movieCount, showCount, seatCount int
	dbConn.QueryRow("SELECT COUNT(*) FROM movies").Scan(&movieCount)
	dbConn.QueryRow("SELECT COUNT(*) FROM shows").Scan(&showCount)
	dbConn.QueryRow("SELECT COUNT(*) FROM seats").Scan(&seatCount)

	log.Printf("Current table status - Movies: %d, Shows: %d, Seats: %d", movieCount, showCount, seatCount)

	// Clear existing data if we have seats already
	if seatCount > 0 {
		log.Println("Clearing existing seats...")
		_, err := dbConn.Exec("DELETE FROM seats")
		if err != nil {
			log.Printf("Error clearing seats: %v", err)
		}
		dbConn.QueryRow("SELECT COUNT(*) FROM seats").Scan(&seatCount)
		log.Printf("Seats after clearing: %d", seatCount)
	}

	// Clear shows if any exist
	if showCount > 0 {
		log.Println("Clearing existing shows...")
		_, err := dbConn.Exec("DELETE FROM shows")
		if err != nil {
			log.Printf("Error clearing shows: %v", err)
		}
		dbConn.QueryRow("SELECT COUNT(*) FROM shows").Scan(&showCount)
		log.Printf("Shows after clearing: %d", showCount)
	}

	// Clear movies if any exist
	if movieCount > 0 {
		log.Println("Clearing existing movies...")
		_, err := dbConn.Exec("DELETE FROM movies")
		if err != nil {
			log.Printf("Error clearing movies: %v", err)
		}
		dbConn.QueryRow("SELECT COUNT(*) FROM movies").Scan(&movieCount)
		log.Printf("Movies after clearing: %d", movieCount)
	}

	// Add movies
	_, err := dbConn.Exec(`
		INSERT INTO movies (title, description, duration, rating, poster_url) VALUES
		('The Dark Knight', 'When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.', 152, 'PG-13', 'https://example.com/dark_knight.jpg'),
		('Inception', 'A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.', 148, 'PG-13', 'https://example.com/inception.jpg'),
		('The Shawshank Redemption', 'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 142, 'R', 'https://example.com/shawshank.jpg')
	`)
	if err != nil {
		log.Printf("Error adding movies: %v", err)
	} else {
		log.Println("Added movies successfully")
	}

	// Add shows
	_, err = dbConn.Exec(`
		INSERT INTO shows (movie_id, screen, start_time, end_time, price) VALUES
		(1, 'Screen 1', '2024-03-20 14:00:00', '2024-03-20 16:32:00', 12.99),
		(1, 'Screen 2', '2024-03-20 18:00:00', '2024-03-20 20:32:00', 14.99),
		(2, 'Screen 1', '2024-03-20 15:00:00', '2024-03-20 17:28:00', 12.99),
		(2, 'Screen 3', '2024-03-20 19:00:00', '2024-03-20 21:28:00', 14.99),
		(3, 'Screen 2', '2024-03-20 16:00:00', '2024-03-20 18:22:00', 12.99),
		(3, 'Screen 1', '2024-03-20 20:00:00', '2024-03-20 22:22:00', 14.99)
	`)
	if err != nil {
		log.Printf("Error adding shows: %v", err)
	} else {
		log.Println("Added shows successfully")
	}

	// Get all show IDs
	rows, err := dbConn.Query("SELECT id FROM shows")
	if err != nil {
		log.Printf("Error getting show IDs: %v", err)
		return
	}
	defer rows.Close()

	var showIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning show ID: %v", err)
			continue
		}
		showIDs = append(showIDs, id)
	}

	log.Printf("Found %d shows to add seats for", len(showIDs))

	// Add seats for each show
	for _, showID := range showIDs {
		log.Printf("Adding seats for show ID: %d", showID)
		// Create 50 seats (5 rows x 10 seats)
		for _, row := range []string{"A", "B", "C", "D", "E"} {
			for seatNum := 1; seatNum <= 10; seatNum++ {
				_, err := dbConn.Exec(`
					INSERT INTO seats (show_id, row_name, seat_number, status)
					VALUES (?, ?, ?, 'available')
				`, showID, row, seatNum)
				if err != nil {
					log.Printf("Error adding seat %s%d for show %d: %v", row, seatNum, showID, err)
				} else {
					log.Printf("Added seat %s%d for show %d", row, seatNum, showID)
				}
			}
		}
		log.Printf("Added seats for show %d", showID)
	}

	// Verify final counts
	dbConn.QueryRow("SELECT COUNT(*) FROM movies").Scan(&movieCount)
	dbConn.QueryRow("SELECT COUNT(*) FROM shows").Scan(&showCount)
	dbConn.QueryRow("SELECT COUNT(*) FROM seats").Scan(&seatCount)

	log.Printf("Final table status - Movies: %d, Shows: %d, Seats: %d", movieCount, showCount, seatCount)
	log.Println("Database seeding completed!")
}

func main() {
	// Parse command line flags
	seed := flag.Bool("seed", false, "Seed the database with sample data")
	flag.Parse()

	// Initialize database
	initDB()
	defer dbConn.Close()

	// If seed flag is provided, seed the database and exit
	if *seed {
		seedDB()
		return
	}

	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Frontend routes
	r.HandleFunc("/", serveHome).Methods("GET")
	r.HandleFunc("/movies", serveMovies).Methods("GET")
	r.HandleFunc("/movies/details", serveMovieDetails).Methods("GET")
	r.HandleFunc("/show", serveShowDetails).Methods("GET")
	r.HandleFunc("/booking/{id}", serveBooking).Methods("GET")

	// API routes
	r.HandleFunc("/api/movies/all", getAllMovies).Methods("GET")
	r.HandleFunc("/api/movies/details", getMovie).Methods("GET")
	r.HandleFunc("/api/movies/shows", getShows).Methods("GET")
	r.HandleFunc("/api/shows/{id}", getShow).Methods("GET")
	r.HandleFunc("/api/shows/{id}/seats", getSeats).Methods("GET")
	r.HandleFunc("/api/bookings", createBooking).Methods("POST")
	r.HandleFunc("/api/bookings/{id}", getBooking).Methods("GET")

	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func serveMovies(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/movies.html"))
	tmpl.Execute(w, nil)
}

func serveMovieDetails(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/movie_details.html"))
	tmpl.Execute(w, nil)
}

func serveShowDetails(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/show_details.html"))
	tmpl.Execute(w, nil)
}

func getAllMovies(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting all movies")

	// Test database connection
	err := dbConn.Ping()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	rows, err := dbConn.Query(`
		SELECT id, title, description, duration, rating, poster_url 
		FROM movies
	`)
	if err != nil {
		log.Printf("Error fetching movies: %v", err)
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.Duration, &movie.Rating, &movie.PosterURL)
		if err != nil {
			log.Printf("Error scanning movie: %v", err)
			http.Error(w, "Failed to scan movie", http.StatusInternalServerError)
			return
		}
		movies = append(movies, movie)
	}

	log.Printf("Found %d movies", len(movies))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	var movieID string
	vars := mux.Vars(r)
	if id := vars["id"]; id != "" {
		movieID = id
	} else {
		movieID = r.URL.Query().Get("id")
	}

	if movieID == "" {
		http.Error(w, "Movie ID not provided", http.StatusBadRequest)
		return
	}

	var movie Movie
	err := dbConn.QueryRow(`
		SELECT id, title, description, duration, rating, poster_url 
		FROM movies 
		WHERE id = ?
	`, movieID).Scan(&movie.ID, &movie.Title, &movie.Description, &movie.Duration, &movie.Rating, &movie.PosterURL)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func getShows(w http.ResponseWriter, r *http.Request) {
	var movieID string
	vars := mux.Vars(r)
	if id := vars["id"]; id != "" {
		movieID = id
	} else {
		movieID = r.URL.Query().Get("id")
	}

	log.Printf("Getting shows for movie ID: %s", movieID)

	if movieID == "" {
		log.Printf("No movie ID provided")
		http.Error(w, "Movie ID not provided", http.StatusBadRequest)
		return
	}

	// Test database connection
	err := dbConn.Ping()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	// First, check if the movie exists
	var movieExists bool
	err = dbConn.QueryRow("SELECT EXISTS(SELECT 1 FROM movies WHERE id = ?)", movieID).Scan(&movieExists)
	if err != nil {
		log.Printf("Error checking movie existence: %v", err)
		http.Error(w, "Error checking movie", http.StatusInternalServerError)
		return
	}

	if !movieExists {
		log.Printf("Movie with ID %s does not exist", movieID)
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	log.Printf("Executing shows query for movie ID: %s", movieID)
	rows, err := dbConn.Query(`
		SELECT s.id, s.movie_id, s.screen, s.start_time, s.end_time, s.price, m.duration
		FROM shows s
		JOIN movies m ON s.movie_id = m.id
		WHERE s.movie_id = ?
	`, movieID)
	if err != nil {
		log.Printf("Error fetching shows: %v", err)
		http.Error(w, "Failed to fetch shows", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var shows []Show
	for rows.Next() {
		var show Show
		err := rows.Scan(&show.ID, &show.MovieID, &show.Screen, &show.StartTime, &show.EndTime, &show.Price, &show.Duration)
		if err != nil {
			log.Printf("Error scanning show: %v", err)
			http.Error(w, "Failed to scan show", http.StatusInternalServerError)
			return
		}
		shows = append(shows, show)
	}

	log.Printf("Found %d shows for movie ID %s", len(shows), movieID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shows)
}

func getShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	showID := vars["id"]

	var show Show
	err := dbConn.QueryRow(`
		SELECT s.id, s.movie_id, s.screen, s.start_time, s.end_time, s.price, m.duration
		FROM shows s
		JOIN movies m ON s.movie_id = m.id
		WHERE s.id = ?
	`, showID).Scan(&show.ID, &show.MovieID, &show.Screen, &show.StartTime, &show.EndTime, &show.Price, &show.Duration)
	if err != nil {
		http.Error(w, "Show not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(show)
}

func getSeats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	showID := vars["id"]
	log.Printf("Getting seats for show ID: %s", showID)

	// First check if the show exists
	var showExists bool
	err := dbConn.QueryRow("SELECT EXISTS(SELECT 1 FROM shows WHERE id = ?)", showID).Scan(&showExists)
	if err != nil {
		log.Printf("Error checking show existence: %v", err)
		http.Error(w, "Error checking show", http.StatusInternalServerError)
		return
	}

	if !showExists {
		log.Printf("Show with ID %s does not exist", showID)
		http.Error(w, "Show not found", http.StatusNotFound)
		return
	}

	// Count seats for this show
	var seatCount int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM seats WHERE show_id = ?", showID).Scan(&seatCount)
	if err != nil {
		log.Printf("Error counting seats: %v", err)
		http.Error(w, "Error counting seats", http.StatusInternalServerError)
		return
	}
	log.Printf("Found %d seats for show ID %s", seatCount, showID)

	// Count how many seats per row (for debugging)
	rows, err := dbConn.Query(`
		SELECT row_name, COUNT(*) as seat_count
		FROM seats 
		WHERE show_id = ?
		GROUP BY row_name
		ORDER BY row_name
	`, showID)
	if err == nil {
		defer rows.Close()
		log.Println("Seats per row:")
		for rows.Next() {
			var row string
			var count int
			if err := rows.Scan(&row, &count); err == nil {
				log.Printf("Row %s: %d seats", row, count)
			}
		}
	}

	// Get all seat details
	rows, err = dbConn.Query(`
		SELECT id, show_id, row_name, seat_number, status 
		FROM seats 
		WHERE show_id = ?
		ORDER BY row_name, seat_number
	`, showID)
	if err != nil {
		log.Printf("Error fetching seats: %v", err)
		http.Error(w, "Failed to fetch seats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seats []Seat
	rowSeatCounts := make(map[string]int) // To track number of seats per row

	for rows.Next() {
		var seat Seat
		err := rows.Scan(&seat.ID, &seat.ShowID, &seat.Row, &seat.SeatNumber, &seat.Status)
		if err != nil {
			log.Printf("Error scanning seat: %v", err)
			http.Error(w, "Failed to scan seat", http.StatusInternalServerError)
			return
		}

		// Count seats per row for debugging
		rowSeatCounts[seat.Row]++

		seats = append(seats, seat)
		log.Printf("Scanned seat: ID=%d, ShowID=%d, Row=%s, Number=%d, Status=%s",
			seat.ID, seat.ShowID, seat.Row, seat.SeatNumber, seat.Status)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning seats: %v", err)
		http.Error(w, "Error after scanning seats", http.StatusInternalServerError)
		return
	}

	// Log seat counts by row for debugging
	log.Println("Final seat counts per row:")
	for row, count := range rowSeatCounts {
		log.Printf("Row %s: %d seats", row, count)
	}

	log.Printf("Successfully retrieved %d seats", len(seats))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seats)
}

// Add booking template handler
func serveBooking(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/booking_confirmation.html"))
	tmpl.Execute(w, nil)
}

// Add createBooking handler
func createBooking(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating new booking")

	// Parse the booking request
	var bookingRequest struct {
		ShowID    int    `json:"show_id"`
		SeatIDs   []int  `json:"seat_ids"`
		UserName  string `json:"user_name"`
		UserEmail string `json:"user_email"`
	}

	err := json.NewDecoder(r.Body).Decode(&bookingRequest)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Printf("Booking request: Show ID=%d, Seats=%v", bookingRequest.ShowID, bookingRequest.SeatIDs)

	// Get show information to calculate price
	var showPrice float64
	err = dbConn.QueryRow("SELECT price FROM shows WHERE id = ?", bookingRequest.ShowID).Scan(&showPrice)
	if err != nil {
		log.Printf("Error fetching show price: %v", err)
		http.Error(w, "Show not found", http.StatusNotFound)
		return
	}

	// Calculate total price
	totalAmount := showPrice * float64(len(bookingRequest.SeatIDs))

	// Begin transaction
	tx, err := dbConn.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if all seats are available
	for _, seatID := range bookingRequest.SeatIDs {
		var status string
		err := tx.QueryRow("SELECT status FROM seats WHERE id = ? AND show_id = ? FOR UPDATE", seatID, bookingRequest.ShowID).Scan(&status)
		if err != nil {
			log.Printf("Error checking seat %d: %v", seatID, err)
			http.Error(w, "Seat not found", http.StatusNotFound)
			return
		}

		if status != "available" {
			log.Printf("Seat %d is not available (status: %s)", seatID, status)
			http.Error(w, "Some seats are not available", http.StatusConflict)
			return
		}
	}

	// Create booking record
	var bookingID int
	result, err := tx.Exec(`
		INSERT INTO bookings (show_id, user_name, user_email, total_amount, booking_time, status)
		VALUES (?, ?, ?, ?, NOW(), 'confirmed')
	`, bookingRequest.ShowID, bookingRequest.UserName, bookingRequest.UserEmail, totalAmount)
	if err != nil {
		log.Printf("Error creating booking: %v", err)
		http.Error(w, "Error creating booking", http.StatusInternalServerError)
		return
	}

	bookingID64, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting booking ID: %v", err)
		http.Error(w, "Error creating booking", http.StatusInternalServerError)
		return
	}
	bookingID = int(bookingID64)

	// Update seat status to booked
	for _, seatID := range bookingRequest.SeatIDs {
		_, err := tx.Exec(`
			UPDATE seats SET status = 'booked', booking_id = ? WHERE id = ?
		`, bookingID, seatID)
		if err != nil {
			log.Printf("Error updating seat %d: %v", seatID, err)
			http.Error(w, "Error updating seat status", http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Booking successfully created. ID: %d", bookingID)

	// Return booking information
	booking := Booking{
		ID:          bookingID,
		ShowID:      bookingRequest.ShowID,
		UserName:    bookingRequest.UserName,
		UserEmail:   bookingRequest.UserEmail,
		SeatIDs:     bookingRequest.SeatIDs,
		TotalAmount: totalAmount,
		Status:      "confirmed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// Add getBooking handler
func getBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID := vars["id"]

	log.Printf("Fetching booking details for ID: %s", bookingID)

	var booking Booking
	err := dbConn.QueryRow(`
		SELECT id, show_id, user_name, user_email, total_amount, booking_time, status
		FROM bookings 
		WHERE id = ?
	`, bookingID).Scan(&booking.ID, &booking.ShowID, &booking.UserName, &booking.UserEmail, &booking.TotalAmount, &booking.BookingTime, &booking.Status)
	if err != nil {
		log.Printf("Error fetching booking: %v", err)
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	// Get seat IDs for this booking
	rows, err := dbConn.Query("SELECT id FROM seats WHERE booking_id = ?", bookingID)
	if err != nil {
		log.Printf("Error fetching seats for booking: %v", err)
		http.Error(w, "Error fetching seats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seatIDs []int
	for rows.Next() {
		var seatID int
		if err := rows.Scan(&seatID); err != nil {
			log.Printf("Error scanning seat ID: %v", err)
			continue
		}
		seatIDs = append(seatIDs, seatID)
	}
	booking.SeatIDs = seatIDs

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}
