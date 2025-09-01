-- name: GetUser :one
SELECT id, username, password_hash FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateRoom :one
INSERT INTO rooms (id, name, owner_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetRoomByID :one
SELECT * FROM rooms WHERE id = $1;

-- name: GetRooms :many
SELECT * FROM rooms ORDER BY created_at DESC;

-- name: UpdateRoom :one
UPDATE rooms SET name = $2 WHERE id = $1 RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms WHERE id = $1;
