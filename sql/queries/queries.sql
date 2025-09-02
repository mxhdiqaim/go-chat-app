-- name: CreateUser :one
INSERT INTO users (id, username, password) VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: UpdateUser :one
UPDATE users SET username = $2, password = $3 WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CreateRoom :one
INSERT INTO rooms (id, name, owner_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetRooms :many
SELECT * FROM rooms ORDER BY created_at DESC;

-- name: GetRoomByID :one
SELECT * FROM rooms WHERE id = $1;

-- name: UpdateRoom :one
UPDATE rooms SET name = $2 WHERE id = $1 RETURNING *;

-- name: DeleteRoom :exec
DELETE FROM rooms WHERE id = $1;

-- name: SearchUsers :many
SELECT * FROM users WHERE username ILIKE $1;

-- name: AddRoomMember :exec
INSERT INTO room_members (room_id, user_id) VALUES ($1, $2);

-- name: RemoveRoomMember :exec
DELETE FROM room_members WHERE room_id = $1 AND user_id = $2;

-- name: IsRoomMember :one
SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2);

-- name: GetRoomMembers :many
SELECT u.id, u.username FROM users AS u JOIN room_members AS rm ON u.id = rm.user_id WHERE rm.room_id = $1;