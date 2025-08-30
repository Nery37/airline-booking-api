CREATE TABLE seats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    flight_id BIGINT NOT NULL,
    seat_no VARCHAR(10) NOT NULL,
    class VARCHAR(20) NOT NULL DEFAULT 'economy',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE,
    UNIQUE KEY uk_flight_seat (flight_id, seat_no),
    INDEX idx_seats_flight (flight_id)
);
