// Global variables
let selectedSeats = new Set();
let currentShowId = null;

// DOM Content Loaded
document.addEventListener('DOMContentLoaded', () => {
    loadMovies();
    setupEventListeners();
});

// Event Listeners
function setupEventListeners() {
    const loginBtn = document.getElementById('loginBtn');
    if (loginBtn) {
        loginBtn.addEventListener('click', handleLogin);
    }
}

// API Functions
async function loadMovies() {
    try {
        const response = await fetch('/api/movies/all');
        if (!response.ok) {
            throw new Error('Failed to load movies');
        }
        const movies = await response.json();
        displayMovies(movies);
    } catch (error) {
        console.error('Error loading movies:', error);
        showError('Failed to load movies. Please try again later.');
    }
}

async function loadShows(movieId) {
    try {
        const response = await fetch(`/api/movies/shows?id=${movieId}`);
        if (!response.ok) {
            throw new Error('Failed to load shows');
        }
        const shows = await response.json();
        displayShows(shows);
    } catch (error) {
        console.error('Error loading shows:', error);
        showError('Failed to load shows. Please try again later.');
    }
}

async function loadSeats(showId) {
    try {
        const response = await fetch(`/api/shows/${showId}/seats`);
        const seats = await response.json();
        displaySeats(seats);
        currentShowId = showId;
    } catch (error) {
        console.error('Error loading seats:', error);
        showError('Failed to load seats. Please try again later.');
    }
}

// Display Functions
function displayMovies(movies) {
    const container = document.getElementById('movies-container');
    if (!container) return;

    container.innerHTML = movies.map(movie => `
        <div class="movie-card" data-movie-id="${movie.id}">
            <a href="/movies/details?id=${movie.id}" class="movie-poster">
                <img src="${movie.poster_url}" alt="${movie.title}">
            </a>
            <div class="movie-info">
                <a href="/movies/details?id=${movie.id}" class="movie-title">
                    <h3>${movie.title}</h3>
                </a>
                <p>${movie.description}</p>
                <button onclick="window.location.href='/movies/details?id=${movie.id}'">View Details</button>
            </div>
        </div>
    `).join('');
}

function displayShows(shows) {
    const container = document.getElementById('shows-container');
    if (!container) return;

    container.style.display = 'block'; // Make shows container visible

    container.innerHTML = shows.map(show => `
        <div class="show-card" data-show-id="${show.id}">
            <div class="show-info">
                <p><strong>Time:</strong> ${new Date(show.start_time).toLocaleString()}</p>
                <p><strong>Screen:</strong> ${show.screen}</p>
                <p><strong>Price:</strong> $${show.price.toFixed(2)}</p>
            </div>
            <div class="show-actions">
                <a href="/show?id=${show.id}" class="btn">View Show</a>
            </div>
        </div>
    `).join('');
}

function displaySeats(seats) {
    const container = document.getElementById('seats-container');
    if (!container) return;

    container.innerHTML = seats.map(seat => `
        <div class="seat ${seat.status}" 
             data-seat-id="${seat.id}"
             onclick="toggleSeat(${seat.id})">
            ${seat.row}${seat.number}
        </div>
    `).join('');
}

// Seat Selection
function toggleSeat(seatId) {
    const seat = document.querySelector(`[data-seat-id="${seatId}"]`);
    if (!seat || seat.classList.contains('booked')) return;

    if (selectedSeats.has(seatId)) {
        selectedSeats.delete(seatId);
        seat.classList.remove('selected');
    } else {
        selectedSeats.add(seatId);
        seat.classList.add('selected');
    }
}

// Booking
async function createBooking() {
    if (!currentShowId || selectedSeats.size === 0) {
        showError('Please select at least one seat');
        return;
    }

    try {
        const response = await fetch('/api/bookings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                show_id: currentShowId,
                seat_ids: Array.from(selectedSeats),
                user_id: 1 // This should be replaced with actual user ID
            })
        });

        if (!response.ok) {
            throw new Error('Booking failed');
        }

        const booking = await response.json();
        showSuccess('Booking successful!');
        // Redirect to booking confirmation page
        window.location.href = `/booking/${booking.id}`;
    } catch (error) {
        console.error('Error creating booking:', error);
        showError('Failed to create booking. Please try again.');
    }
}

// Utility Functions
function showError(message) {
    // Implement error notification
    alert(message);
}

function showSuccess(message) {
    // Implement success notification
    alert(message);
}

function handleLogin() {
    // Implement login functionality
    alert('Login functionality to be implemented');
} 