package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/leophys/userz"
	"github.com/leophys/userz/store/pg/postgres"
)

var _ userz.Store = &PGStore{}

// PGStore is the implementation of the store with a postgresql backend.
type PGStore struct {
	db
	q *postgres.Queries
}

func NewPGStore(ctx context.Context, databaseURL string) (userz.Store, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &PGStore{
		db: &PGPooledConn{pool},
		q:  postgres.New(pool),
	}, nil
}

func (s *PGStore) Add(ctx context.Context, user *userz.UserData) (*userz.User, error) {
	params := postgres.AddParams{
		Nickname: user.NickName,
		Password: user.Password.String(),
		Email:    user.Email,
	}

	if user.FirstName != "" {
		params.FirstName = sql.NullString{
			String: user.FirstName,
			Valid:  true,
		}
	}

	if user.LastName != "" {
		params.LastName = sql.NullString{
			String: user.LastName,
			Valid:  true,
		}
	}

	if user.Country != "" {
		params.Country = sql.NullString{
			String: user.Country,
			Valid:  true,
		}
	}

	pgResult, err := s.q.Add(ctx, params)
	if err != nil {
		return nil, err
	}

	password := userz.Password(pgResult.Password)

	result := &userz.User{
		Id:        pgResult.ID.String(),
		FirstName: pgResult.FirstName.String,
		LastName:  pgResult.LastName.String,
		NickName:  pgResult.Nickname,
		Password:  &password,
		Email:     pgResult.Email,
		Country:   pgResult.Country.String,
		CreatedAt: pgResult.CreatedAt.Time,
		UpdatedAt: pgResult.UpdatedAt.Time,
	}

	return result, nil
}

func (s *PGStore) Update(ctx context.Context, id string, user *userz.UserData) (*userz.User, error) {
	uuidId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	params := postgres.UpdateParams{
		ID: uuidId,
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx.(pgx.Tx))

	cur, err := q.Get(ctx, uuidId)
	if err != nil {
		return nil, err
	}

	if user.FirstName != "" {
		params.FirstName = sql.NullString{
			String: user.FirstName,
			Valid:  true,
		}
	} else {
		params.FirstName = cur.FirstName
	}

	if user.LastName != "" {
		params.LastName = sql.NullString{
			String: user.LastName,
			Valid:  true,
		}
	} else {
		params.LastName = cur.LastName
	}

	if user.NickName != "" {
		params.Nickname = user.NickName
	} else {
		params.Nickname = cur.Nickname
	}

	if user.Password != nil {
		params.Password = user.Password.String()
	} else {
		params.Password = cur.Password
	}

	if user.Email != "" {
		params.Email = user.Email
	} else {
		params.Email = cur.Email
	}

	if user.Country != "" {
		params.Country = sql.NullString{
			String: user.Country,
			Valid:  true,
		}
	} else {
		params.Country = cur.Country
	}

	pgResult, err := q.Update(ctx, params)
	if err != nil {
		return nil, err
	}

	password := userz.Password(pgResult.Password)

	result := &userz.User{
		Id:        pgResult.ID.String(),
		FirstName: pgResult.FirstName.String,
		LastName:  pgResult.LastName.String,
		NickName:  pgResult.Nickname,
		Password:  &password,
		Email:     pgResult.Email,
		Country:   pgResult.Country.String,
		CreatedAt: pgResult.CreatedAt.Time,
		UpdatedAt: pgResult.UpdatedAt.Time,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PGStore) Remove(ctx context.Context, id string) (*userz.User, error) {
	uuidId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	pgResult, err := s.q.Remove(ctx, uuidId)
	if err != nil {
		return nil, err
	}

	password := userz.Password(pgResult.Password)

	result := &userz.User{
		Id:        pgResult.ID.String(),
		FirstName: pgResult.FirstName.String,
		LastName:  pgResult.LastName.String,
		NickName:  pgResult.Nickname,
		Password:  &password,
		Email:     pgResult.Email,
		Country:   pgResult.Country.String,
		CreatedAt: pgResult.CreatedAt.Time,
		UpdatedAt: pgResult.UpdatedAt.Time,
	}

	return result, nil
}

func (s *PGStore) List(ctx context.Context, filter *userz.Filter, pageSize uint) (userz.Iterator[[]*userz.User], error) {
	filterStr, err := formatFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize filter into statement: %w", err)
	}

	var filterHash string
	if filter != nil {
		filterHash, err = filter.Hash()
		if err != nil {
			return nil, fmt.Errorf("failed to get hash of filter: %w", err)
		}
	}

	query, err := prepareListPaginated(ctx, s.db, preparePaginatedParams{
		queryName: filterHash,
		filter:    filterStr,
		pageSize:  pageSize,
		orderBy:   userz.OrdByCreatedAt,
	})

	return &PGIterator{
		pageSize: pageSize,
		dbtx:     s.db,
		query:    query,
	}, nil
}

func (s *PGStore) Page(ctx context.Context, filter *userz.Filter, params *userz.PageParams) ([]*userz.User, error) {
	filterStr, err := formatFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize filter into statement: %w", err)
	}

	var filterHash string
	if filter != nil {
		filterHash, err = filter.Hash()
		if err != nil {
			return nil, fmt.Errorf("failed to get hash of filter: %w", err)
		}
	}

	query, err := prepareListPaginated(ctx, s.db, preparePaginatedParams{
		queryName: filterHash,
		filter:    filterStr,
		pageSize:  params.Size,
		orderBy:   params.Order,
	})
	if err != nil {
		return nil, err
	}

	users, _, err := query(ctx, params.Offset)
	if err != nil {
		return nil, err
	}
	return users, nil
}
