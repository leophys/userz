package pg

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leophys/userz"
	"github.com/leophys/userz/store/pg/postgres"
)

const (
	add = `-- name: Add :one
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
	update = `-- name: Update :one
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
	get = `-- name: Get :one
SELECT id, first_name, last_name, nickname, password, email, country, created_at, updated_at
FROM users
WHERE id = $1
`
	remove = `-- name: Remove :one
DELETE FROM users
WHERE
    id = $1
RETURNING id, first_name, last_name, nickname, password, email, country, created_at, updated_at
`
)

func TestStoreAdd(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	id := "e3a190a2-e22e-460e-80dc-1af731744031"
	password, err := userz.NewPassword("1234567890")
	require.NoError(err)
	createdAtStr := "2022-11-27T12:22:05Z"
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	require.NoError(err)

	user := userz.User{
		Id:        id,
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "JD",
		Password:  password,
		Email:     "jd@example.com",
		Country:   "US",
		CreatedAt: createdAt,
	}
	row := userRow(user)

	fakeDB := &mockDB{
		queryRow: map[string]pgx.Row{
			fmtSql(add,
				"John",
				"Doe",
				"JD",
				password.String(),
				"jd@example.com",
				"US",
			): &row,
		},
	}

	store := &PGStore{
		db: fakeDB,
		q:  postgres.New(fakeDB),
	}

	res, err := store.Add(context.TODO(), &userz.UserData{
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "JD",
		Password:  password,
		Email:     "jd@example.com",
		Country:   "US",
	})
	assert.NoError(err)
	require.NotNil(res)
	assert.Equal(user, *res)
}

func TestStoreUpdate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	id := "e3a190a2-e22e-460e-80dc-1af731744031"
	password, err := userz.NewPassword("1234567890")
	require.NoError(err)
	updatedAtStr := "2022-11-27T12:22:05Z"
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	require.NoError(err)

	user := userz.User{
		Id:        id,
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "JD",
		Password:  password,
		Email:     "jd@example.com",
		Country:   "US",
		UpdatedAt: updatedAt,
	}
	row := userRow(user)

	fakeDB := &mockDB{
		queryRow: map[string]pgx.Row{
			fmtSql(update,
				id,
				"John",
				"Doe",
				"JD",
				password.String(),
				"jd@example.com",
				"US",
			): &row,
			fmtSql(get, id): &row,
		},
	}

	store := &PGStore{
		db: fakeDB,
		q:  postgres.New(fakeDB),
	}

	res, err := store.Update(context.TODO(), id, &userz.UserData{
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "JD",
		Password:  password,
		Email:     "jd@example.com",
		Country:   "US",
	})
	assert.NoError(err)
	require.NotNil(res)
	assert.Equal(user, *res)
}

func TestStoreRemove(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	id := "e3a190a2-e22e-460e-80dc-1af731744031"
	password, err := userz.NewPassword("1234567890")
	require.NoError(err)
	createdAtStr := "2022-11-27T12:22:05Z"
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	require.NoError(err)
	updatedAtStr := "2022-11-27T12:22:05Z"
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	require.NoError(err)

	user := userz.User{
		Id:        id,
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "JD",
		Password:  password,
		Email:     "jd@example.com",
		Country:   "US",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	row := userRow(user)

	fakeDB := &mockDB{
		queryRow: map[string]pgx.Row{
			fmtSql(remove, id): &row,
		},
	}

	store := &PGStore{
		db: fakeDB,
		q:  postgres.New(fakeDB),
	}

	res, err := store.Remove(context.TODO(), id)
	assert.NoError(err)
	require.NotNil(res)
	assert.Equal(user, *res)
}
