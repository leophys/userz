// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: queries.sql

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const add = `-- name: Add :one
INSERT INTO users (
    first_name,
    last_name,
    nickname,
    password,
    email,
    country,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id, first_name, last_name, nickname, password, email, country, created_at, updated_at
`

type AddParams struct {
	FirstName sql.NullString
	LastName  sql.NullString
	Nickname  string
	Password  []byte
	Email     string
	Country   sql.NullString
}

func (q *Queries) Add(ctx context.Context, arg AddParams) (User, error) {
	row := q.db.QueryRow(ctx, add,
		arg.FirstName,
		arg.LastName,
		arg.Nickname,
		arg.Password,
		arg.Email,
		arg.Country,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Nickname,
		&i.Password,
		&i.Email,
		&i.Country,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const get = `-- name: Get :one
SELECT id, first_name, last_name, nickname, password, email, country, created_at, updated_at
FROM users
WHERE id = $1
`

func (q *Queries) Get(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, get, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Nickname,
		&i.Password,
		&i.Email,
		&i.Country,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const remove = `-- name: Remove :one
DELETE FROM users
WHERE
    id = $1
RETURNING id, first_name, last_name, nickname, password, email, country, created_at, updated_at
`

func (q *Queries) Remove(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, remove, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Nickname,
		&i.Password,
		&i.Email,
		&i.Country,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const update = `-- name: Update :one
UPDATE users SET
    first_name = $2,
    last_name = $3,
    nickname = $4,
    password = $5,
    email = $6,
    country = $7,
    updated_at = NOW()
WHERE
    id = $1
RETURNING id, first_name, last_name, nickname, password, email, country, created_at, updated_at
`

type UpdateParams struct {
	ID        uuid.UUID
	FirstName sql.NullString
	LastName  sql.NullString
	Nickname  string
	Password  []byte
	Email     string
	Country   sql.NullString
}

func (q *Queries) Update(ctx context.Context, arg UpdateParams) (User, error) {
	row := q.db.QueryRow(ctx, update,
		arg.ID,
		arg.FirstName,
		arg.LastName,
		arg.Nickname,
		arg.Password,
		arg.Email,
		arg.Country,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Nickname,
		&i.Password,
		&i.Email,
		&i.Country,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
