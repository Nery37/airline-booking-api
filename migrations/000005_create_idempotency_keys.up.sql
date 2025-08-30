CREATE TABLE idempotency_keys (
    request_id VARCHAR(100) NOT NULL,
    route VARCHAR(100) NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    response_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (request_id, route),
    INDEX idx_idempotency_user_route (user_id, route),
    INDEX idx_idempotency_created (created_at)
);
