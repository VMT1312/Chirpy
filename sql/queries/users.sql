-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, chirpy_red;

-- name: ResetUser :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdatePasswordByID :one
UPDATE users
SET hashed_password = $1, updated_at = NOW(), email = $2
WHERE id = $3
RETURNING id, created_at, updated_at, email, chirpy_red;

-- name: UpgradeUserByID :exec
UPDATE users
SET chirpy_red = TRUE, updated_at = NOW()
WHERE id = $1;