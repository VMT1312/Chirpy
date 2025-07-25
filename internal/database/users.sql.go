// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: users.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES(
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, chirpy_red
`

type CreateUserParams struct {
	Email          string
	HashedPassword string
}

type CreateUserRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	ChirpyRed bool
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.HashedPassword)
	var i CreateUserRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.ChirpyRed,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, hashed_password, chirpy_red FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.ChirpyRed,
	)
	return i, err
}

const resetUser = `-- name: ResetUser :exec
DELETE FROM users
`

func (q *Queries) ResetUser(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, resetUser)
	return err
}

const updatePasswordByID = `-- name: UpdatePasswordByID :one
UPDATE users
SET hashed_password = $1, updated_at = NOW(), email = $2
WHERE id = $3
RETURNING id, created_at, updated_at, email, chirpy_red
`

type UpdatePasswordByIDParams struct {
	HashedPassword string
	Email          string
	ID             uuid.UUID
}

type UpdatePasswordByIDRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	ChirpyRed bool
}

func (q *Queries) UpdatePasswordByID(ctx context.Context, arg UpdatePasswordByIDParams) (UpdatePasswordByIDRow, error) {
	row := q.db.QueryRowContext(ctx, updatePasswordByID, arg.HashedPassword, arg.Email, arg.ID)
	var i UpdatePasswordByIDRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.ChirpyRed,
	)
	return i, err
}

const upgradeUserByID = `-- name: UpgradeUserByID :exec
UPDATE users
SET chirpy_red = TRUE, updated_at = NOW()
WHERE id = $1
`

func (q *Queries) UpgradeUserByID(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, upgradeUserByID, id)
	return err
}
