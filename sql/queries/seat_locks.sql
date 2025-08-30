-- name: GetSeatLock :one
SELECT * FROM seat_locks WHERE flight_id = ? AND seat_no = ?;

-- name: CreateSeatLock :exec
INSERT INTO seat_locks (flight_id, seat_no, holder_id, expires_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    holder_id = VALUES(holder_id),
    expires_at = VALUES(expires_at),
    updated_at = CURRENT_TIMESTAMP;

-- name: UpdateSeatLock :execrows
UPDATE seat_locks 
SET holder_id = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP
WHERE flight_id = ? AND seat_no = ? 
AND (expires_at < NOW() OR holder_id = ?);

-- name: ConfirmSeatLock :execrows
UPDATE seat_locks 
SET expires_at = NULL, updated_at = CURRENT_TIMESTAMP
WHERE flight_id = ? AND seat_no = ? AND holder_id = ? AND expires_at > NOW();

-- name: ReleaseSeatLock :exec
DELETE FROM seat_locks 
WHERE flight_id = ? AND seat_no = ? AND holder_id = ?;

-- name: CleanupExpiredLocks :exec
DELETE FROM seat_locks WHERE expires_at < NOW();

-- name: ListFlightSeatLocks :many
SELECT flight_id, seat_no, holder_id, expires_at, created_at, updated_at
FROM seat_locks 
WHERE flight_id = ?;
