<<<<<<< HEAD
# Cinema Ticket Booking System

A modern, efficient cinema ticket booking system built with Go, MySQL, and a responsive web interface.

## Features

- Real-time seat booking with concurrency management
- User-friendly interface for browsing movies and shows
- Secure booking process with transaction management
- Responsive design for all devices
- JSON-based data exchange
- Comprehensive error handling
- Unit testing for critical functionalities

## Prerequisites

- Go 1.21 or later
- MySQL 8.0 or later
- Git

## Setup Instructions

1. Clone the repository:
```bash
git clone <repository-url>
cd cinema-ticket-booking
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
mysql -u root -p
CREATE DATABASE cinema_booking;
```

4. Configure environment variables:
Create a `.env` file in the root directory with the following content:
```
DB_USER=your_mysql_username
DB_PASSWORD=your_mysql_password
DB_HOST=localhost:3306
DB_NAME=cinema_booking
```

5. Run the application:
```bash
go run main.go
```

The application will be available at `http://localhost:8080`

## Project Structure

```
cinema-ticket-booking/
├── main.go              # Main application entry point
├── models/              # Database models
├── handlers/            # HTTP request handlers
├── db/                  # Database connection and configuration
├── static/              # Static files (CSS, JS, images)
│   ├── css/
│   └── js/
├── templates/           # HTML templates
└── tests/              # Unit tests
```

## API Endpoints

- `GET /api/movies` - Get all movies
- `GET /api/movies/{id}` - Get movie details
- `GET /api/movies/{id}/shows` - Get shows for a movie
- `GET /api/shows/{id}/seats` - Get seats for a show
- `POST /api/bookings` - Create a new booking
- `GET /api/bookings/{id}` - Get booking details

## Testing

Run the test suite:
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
=======
# CinemaTicketBooking-System
Created a Cinema ticket booking system using Go, HTML, CSS &amp; MySQL
>>>>>>
