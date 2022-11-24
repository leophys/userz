package pg

import (
	"context"
	"sync"

	"github.com/leophys/userz"
	"github.com/leophys/userz/store/pg/postgres"
)

var _ userz.Iterator[[]*userz.User] = &PGIterator{}

type queryFunc func(ctx context.Context, offset uint) ([]*userz.User, uint, error)

type PGIterator struct {
	pageSize  uint
	totalRows uint
	dbtx      postgres.DBTX
	offset    uint
	query     queryFunc
	mu        sync.Mutex
}

func (i *PGIterator) Len() userz.PaginationData {
	return userz.PaginationData{
		TotalElements: i.totalRows,
		TotalPages:    i.totalRows/i.pageSize + 1,
		PageSize:      i.pageSize,
	}
}

func (i *PGIterator) Next(ctx context.Context) ([]*userz.User, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	result, rows, err := i.query(ctx, i.offset)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, userz.ErrNoMorePages
	}

	i.offset += uint(len(result))
	i.totalRows = rows

	return result, nil
}
