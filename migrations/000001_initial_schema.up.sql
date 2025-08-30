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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_flights_route_date (origin, destination, departure_time),
    INDEX idx_flights_airline (airline),
    INDEX idx_flights_departure (departure_time)
);
