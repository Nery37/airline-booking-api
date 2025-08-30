-- name: GetTicket :one
SELECT * FROM tickets WHERE id = ?;

-- name: GetTicketByPNR :one
SELECT * FROM tickets WHERE pnr_code = ?;

-- name: GetTicketByFlightSeat :one
SELECT * FROM tickets WHERE flight_id = ? AND seat_no = ?;

-- name: CreateTicket :execlastid
INSERT INTO tickets (flight_id, seat_no, user_id, price_amount, currency, pnr_code, payment_ref)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListUserTickets :many
SELECT * FROM tickets WHERE user_id = ? ORDER BY created_at DESC;

-- name: ListFlightTickets :many
SELECT * FROM tickets WHERE flight_id = ? ORDER BY seat_no;
