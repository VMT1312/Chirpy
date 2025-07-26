-- +goose Up
ALTER TABLE users
ADD COLUMN chirpy_red BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
DROP COLUMN chirpy_red;