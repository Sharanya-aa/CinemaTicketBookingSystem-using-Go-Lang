-- First, let's add some sample movies if they don't exist
INSERT INTO movies (title, description, duration, rating, poster_url) VALUES
('The Dark Knight', 'When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.', 152, 'PG-13', 'https://example.com/dark_knight.jpg'),
('Inception', 'A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.', 148, 'PG-13', 'https://example.com/inception.jpg'),
('The Shawshank Redemption', 'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 142, 'R', 'https://example.com/shawshank.jpg');

-- Now let's add shows for each movie
-- For The Dark Knight (assuming it's movie_id = 1)
INSERT INTO shows (movie_id, screen, start_time, end_time, price) VALUES
(1, 1, '2024-03-20 14:00:00', '2024-03-20 16:32:00', 12.99),
(1, 2, '2024-03-20 18:00:00', '2024-03-20 20:32:00', 14.99);

-- For Inception (assuming it's movie_id = 2)
INSERT INTO shows (movie_id, screen, start_time, end_time, price) VALUES
(2, 1, '2024-03-20 15:00:00', '2024-03-20 17:28:00', 12.99),
(2, 3, '2024-03-20 19:00:00', '2024-03-20 21:28:00', 14.99);

-- For Shawshank Redemption (assuming it's movie_id = 3)
INSERT INTO shows (movie_id, screen, start_time, end_time, price) VALUES
(3, 2, '2024-03-20 16:00:00', '2024-03-20 18:22:00', 12.99),
(3, 1, '2024-03-20 20:00:00', '2024-03-20 22:22:00', 14.99); 