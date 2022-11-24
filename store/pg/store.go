package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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
	params := postgres.UpdateParams{}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	uuidId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	cur, err := s.q.Get(ctx, uuidId)
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

	pgResult, err := s.q.Update(ctx, params)
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

func (s *PGStore) List(ctx context.Context, filter *userz.Filter[any], pageSize uint) (userz.Iterator[[]*userz.User], error) {
	// This is a dirty trick to constrain the type parameter of *userz.Filter to string
	var f *userz.Filter[string]
	var i any = filter
	switch ff := i.(type) {
	case *userz.Filter[string]:
		f = ff
	default:
		return nil, fmt.Errorf("filter type not acceptable: %T", ff)
	}

	filterStr, err := formatFilter(f)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize filter into statement: %w", err)
	}

	query, err := prepareListPaginated(ctx, s.db, preparePaginatedParams{
		queryName: hash(filterStr),
		filter:    filterStr,
		pageSize:  pageSize,
	})

	return &PGIterator{
		pageSize: pageSize,
		dbtx:     s.db,
		query:    query,
	}, nil
}
