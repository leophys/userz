package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/leophys/userz"
)

const listPaginated = `-- name: ListPaginated :many
SELECT
    id, first_name, last_name, nickname, password, email, country, created_at, updated_at,
    count(*) OVER() AS total_elements
FROM users
WHERE %s
ORDER BY %s
OFFSET $1
LIMIT $2
`

type preparePaginatedParams struct {
	queryName string
	filter    string
	pageSize  uint
	orderBy   userz.Order
}

type listPaginatedRow struct {
	ID        uuid.UUID
	FirstName sql.NullString
	LastName  sql.NullString
	Nickname  string
	Password  string
	Email     string
	Country   sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

func prepareListPaginated(ctx context.Context, db db, params preparePaginatedParams) (queryFunc, error) {
	query := fmt.Sprintf(listPaginated, params.filter, params.orderBy)

	if _, err := db.Prepare(
		ctx,
		params.queryName,
		query); err != nil {
		return nil, err
	}

	return func(ctx context.Context, offset uint) ([]*userz.User, uint, error) {
		rows, err := db.Query(ctx, query, offset, params.pageSize)
		if err != nil {
			return nil, 0, err
		}
		defer rows.Close()

		var totalRows int64
		var result []*userz.User

		for rows.Next() {
			var i listPaginatedRow
			if err := rows.Scan(
				&i.ID,
				&i.FirstName,
				&i.LastName,
				&i.Nickname,
				&i.Password,
				&i.Email,
				&i.Country,
				&i.CreatedAt,
				&i.UpdatedAt,
				&totalRows,
			); err != nil {
				return nil, 0, err
			}

			password := userz.Password(i.Password)
			result = append(result, &userz.User{
				Id:        i.ID.String(),
				FirstName: i.FirstName.String,
				LastName:  i.LastName.String,
				NickName:  i.Nickname,
				Password:  &password,
				Email:     i.Email,
				Country:   i.Country.String,
				CreatedAt: i.CreatedAt.Time,
				UpdatedAt: i.UpdatedAt.Time,
			})
		}

		if err := rows.Err(); err != nil {
			return nil, 0, err
		}

		return result, uint(totalRows), nil
	}, nil
}
