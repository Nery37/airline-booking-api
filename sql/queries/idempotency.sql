-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys WHERE request_id = ? AND route = ?;

-- name: CreateIdempotencyKey :exec
INSERT INTO idempotency_keys (request_id, route, user_id, response_hash)
VALUES (?, ?, ?, ?);

-- name: CleanupOldIdempotencyKeys :exec
DELETE FROM idempotency_keys WHERE created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR);
