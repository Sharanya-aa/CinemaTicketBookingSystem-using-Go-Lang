document.addEventListener('DOMContentLoaded', function() {
    // Get show ID from URL
    const urlParams = new URLSearchParams(window.location.search);
    const showId = urlParams.get('id');

    if (showId) {
        loadShowDetails(showId);
    } else {
        alert('No show ID provided');
        window.location.href = '/movies';
    }
});

let selectedSeats = [];
let showPrice = 0;

async function loadShowDetails(showId) {
    try {
        const response = await fetch(`/api/shows/${showId}`);
        if (!response.ok) {
            throw new Error('Failed to load show details');
        }

        const show = await response.json();
        showPrice = show.price;

        // Load movie details
        const movieResponse = await fetch(`/api/movies/${show.movie_id}`);
        if (!movieResponse.ok) {
            throw new Error('Failed to load movie details');
        }

        const movie = await movieResponse.json();

        // Display movie and show details
        displayMovieInfo(movie);
        displayShowInfo(show);

        // Load and display seats
        loadSeats(showId);
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to load show details');
    }
}

function displayMovieInfo(movie) {
    const movieInfoDiv = document.getElementById('movie-info');
    movieInfoDiv.innerHTML = `
        <div class="movie-card">
            <img src="${movie.poster_url}" alt="${movie.title}">
            <div class="movie-info">
                <h1>${movie.title}</h1>
                <p>${movie.description}</p>
                <p><strong>Duration:</strong> ${movie.duration} minutes</p>
                <p><strong>Rating:</strong> ${movie.rating}</p>
            </div>
        </div>
    `;
}

function displayShowInfo(show) {
    const showDetailsDiv = document.getElementById('show-details');
    showDetailsDiv.innerHTML = `
        <div class="show-info-card">
            <p><strong>Screen:</strong> ${show.screen}</p>
            <p><strong>Show Time:</strong> ${new Date(show.start_time).toLocaleString()}</p>
            <p><strong>Duration:</strong> ${show.duration} minutes</p>
            <p><strong>Price per Seat:</strong> $${show.price}</p>
        </div>
    `;
}

async function loadSeats(showId) {
    try {
        const response = await fetch(`/api/shows/${showId}/seats`);
        if (!response.ok) {
            throw new Error('Failed to load seats');
        }

        const seats = await response.json();
        displaySeats(seats);
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to load seats');
    }
}

function displaySeats(seats) {
    const seatsContainer = document.getElementById('seats-container');
    seatsContainer.innerHTML = '';

    // Group seats by row
    const seatsByRow = {};
    seats.forEach(seat => {
        if (!seatsByRow[seat.row]) {
            seatsByRow[seat.row] = [];
        }
        seatsByRow[seat.row].push(seat);
    });

    // Create seat grid
    Object.keys(seatsByRow).sort().forEach(row => {
        const rowDiv = document.createElement('div');
        rowDiv.className = 'seat-row';
        
        seatsByRow[row].forEach(seat => {
            const seatButton = document.createElement('button');
            seatButton.className = `seat ${seat.status}`;
            seatButton.textContent = seat.seat_number;
            seatButton.dataset.seatId = seat.id;
            seatButton.dataset.row = seat.row;
            seatButton.dataset.number = seat.seat_number;
            
            if (seat.status === 'booked') {
                seatButton.disabled = true;
            } else {
                seatButton.addEventListener('click', () => toggleSeatSelection(seatButton, seat));
            }
            
            rowDiv.appendChild(seatButton);
        });
        
        seatsContainer.appendChild(rowDiv);
    });
}

function toggleSeatSelection(button, seat) {
    const seatId = seat.id;
    const seatIndex = selectedSeats.findIndex(s => s.id === seatId);

    if (seatIndex === -1) {
        // Select seat
        selectedSeats.push(seat);
        button.classList.add('selected');
    } else {
        // Deselect seat
        selectedSeats.splice(seatIndex, 1);
        button.classList.remove('selected');
    }

    updateBookingSummary();
}

function updateBookingSummary() {
    const selectedSeatsDiv = document.getElementById('selected-seats');
    const totalSpan = document.getElementById('total');

    if (selectedSeats.length === 0) {
        selectedSeatsDiv.innerHTML = '<p>No seats selected</p>';
        totalSpan.textContent = '0';
        return;
    }

    const seatsList = selectedSeats.map(seat => 
        `${seat.row}${seat.seat_number}`
    ).join(', ');

    selectedSeatsDiv.innerHTML = `<p>${seatsList}</p>`;
    totalSpan.textContent = (selectedSeats.length * showPrice).toFixed(2);
}

async function createBooking() {
    if (selectedSeats.length === 0) {
        alert('Please select at least one seat');
        return;
    }

    try {
        const response = await fetch('/api/bookings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                show_id: new URLSearchParams(window.location.search).get('id'),
                seats: selectedSeats.map(seat => seat.id)
            })
        });

        if (!response.ok) {
            throw new Error('Failed to create booking');
        }

        const booking = await response.json();
        alert('Booking successful!');
        window.location.href = `/bookings/${booking.id}`;
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to create booking');
    }
} 