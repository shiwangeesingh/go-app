-- name: InsertUser :one
INSERT INTO users (name, age, gender, email, password) 
VALUES ($1, $2, $3, $4, $5)
RETURNING id;


-- name: GetUsers :many
SELECT id, name, age, gender, email FROM users;


-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;


-- name: UpdateUser :exec
UPDATE users 
SET name = $2, age = $3, gender = $4, email = $5 
WHERE id = $1;

-- name: InsertGeneration :one
INSERT INTO user_generation (user_id, generation, grade)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserById :exec
SELECT id, name, age, gender, email FROM users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT id, name, age, gender, email, password 
FROM users 
WHERE email = $1;
