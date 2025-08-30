CREATE TABLE tickets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    flight_id BIGINT NOT NULL,
    seat_no VARCHAR(10) NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    price_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pnr_code VARCHAR(10) NOT NULL UNIQUE,
    payment_ref VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE,
    UNIQUE KEY uk_flight_seat_ticket (flight_id, seat_no),
    INDEX idx_tickets_user (user_id),
    INDEX idx_tickets_pnr (pnr_code),
    INDEX idx_tickets_payment (payment_ref)
);
