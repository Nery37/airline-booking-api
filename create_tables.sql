CREATE TABLE flights (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    origin VARCHAR(10) NOT NULL,
    destination VARCHAR(10) NOT NULL,
    departure_time DATETIME NOT NULL,
    arrival_time DATETIME NOT NULL,
    airline VARCHAR(10) NOT NULL,
    aircraft VARCHAR(20) NOT NULL,
    fare_class VARCHAR(20) NOT NULL DEFAULT 'economy',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE seats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    flight_id BIGINT NOT NULL,
    seat_no VARCHAR(10) NOT NULL,
    class VARCHAR(20) NOT NULL DEFAULT 'economy',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE,
    UNIQUE KEY uk_flight_seat (flight_id, seat_no)
);

CREATE TABLE seat_locks (
    flight_id BIGINT NOT NULL,
    seat_no VARCHAR(10) NOT NULL,
    holder_id VARCHAR(100) NOT NULL,
    expires_at DATETIME NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (flight_id, seat_no),
    FOREIGN KEY (flight_id) REFERENCES flights(id) ON DELETE CASCADE
);

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
    UNIQUE KEY uk_flight_seat_ticket (flight_id, seat_no)
);

CREATE TABLE idempotency_keys (
    request_id VARCHAR(100) NOT NULL,
    route VARCHAR(100) NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    response_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (request_id, route)
);
