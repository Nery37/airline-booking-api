CREATE TABLE seat_locks (
    flight_id BIGINT NOT NULL,
    seat_no VARCHAR(10) NOT NULL,
    holder_id VARCHAR(100) NOT NULL,
    expires_at DATETIME NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (flight_id, seat_no),
    FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE,
    INDEX idx_seat_locks_expires (expires_at),
    INDEX idx_seat_locks_holder (holder_id)
);
