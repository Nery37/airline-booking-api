-- name: GetSeat :one
SELECT * FROM seats WHERE flight_id = ? AND seat_no = ?;

-- name: ListSeats :many
SELECT * FROM seats WHERE flight_id = ? ORDER BY seat_no;

-- name: CreateSeat :execlastid
INSERT INTO seats (flight_id, seat_no, class)
VALUES (?, ?, ?);

-- name: CreateSeats :exec
INSERT INTO seats (flight_id, seat_no, class)
VALUES (?, ?, ?);

-- name: DeleteSeats :exec
DELETE FROM seats WHERE flight_id = ?;
