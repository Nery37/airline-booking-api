-- name: GetFlight :one
SELECT * FROM flights WHERE id = ?;

-- name: ListFlights :many
SELECT * FROM flights
WHERE (@origin = '' OR origin = @origin)
AND (@destination = '' OR destination = @destination)
AND (@airline = '' OR airline = @airline)
AND (@date_start = '' OR departure_time >= @date_start)
AND (@date_end = '' OR departure_time <= @date_end)
ORDER BY departure_time
LIMIT ? OFFSET ?;

-- name: CreateFlight :execlastid
INSERT INTO flights (origin, destination, departure_time, arrival_time, airline, aircraft, fare_class)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateFlight :exec
UPDATE flights 
SET origin = ?, destination = ?, departure_time = ?, arrival_time = ?, 
    airline = ?, aircraft = ?, fare_class = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteFlight :exec
DELETE FROM flights WHERE id = ?;
