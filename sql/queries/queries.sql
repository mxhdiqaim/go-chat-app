-- name: GetUser :one
SELECT id, username, password_hash FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- -- name: CreateUser :one
-- INSERT INTO users (id, username) VALUES ($1, $2) RETURNING *;

-- -- name: GetUserByUsername :one
-- SELECT * FROM users WHERE username = $1 LIMIT 1;

-- -- name: CreateRoom :one
-- INSERT INTO rooms (id, name) VALUES ($1, $2) RETURNING *;

-- -- name: GetRoomByName :one
-- SELECT * FROM rooms WHERE name = $1 LIMIT 1;
