package userz

import (
	"context"
	"errors"
	"time"
)

// Store represents the storage backend for the Users.
type Store interface {
	Add(ctx context.Context, user *UserData) (*User, error)
	Update(ctx context.Context, id string, user *UserData) (*User, error)
	Remove(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, filter *Filter[any], pageSize uint) (Iterator[[]*User], error)
}

// UserData represents the data needed to create or alter a user.
type UserData struct {
	FirstName string
	LastName  string
	NickName  string
	Password  *Password
	Email     string
	Country   string
}

// Filter is a condition to be used to filter users. The backend type
// represents the output type a concrete implementation will produce
// as output of the evaluation of the filter.
type Filter[Backend any] struct {
	Id        string
	FirstName Condition[string, Backend]
	LastName  Condition[string, Backend]
	NickName  Condition[string, Backend]
	Email     Condition[string, Backend]
	Country   Condition[string, Backend]
	CreatedAt Condition[time.Time, Backend]
	UpdatedAt Condition[time.Time, Backend]
}

// Condition is the interface any backend will need to implement in order
// to translate the given condition in a valid expression for the backend
// at hand.
type Condition[T Conditionable, Backend any] interface {
	// Evaluate translates the abstract Condition into a form that is
	// usable by the backend.
	Evaluate() (Backend, error)
	// Hash returns a unique identified deterministically derived by the
	// values of the condition.
	Hash() (string, error)
}

// Iterator is the interface to iterate over the results.
type Iterator[T any] interface {
	// Len returns some data regarding the pagination.
	Len() PaginationData
	// Next returns the next page. It returns ErrNoMorePages after the last page.
	Next(ctx context.Context) (T, error)
}

// PaginationData regards the global information pertaining the pagination.
type PaginationData struct {
	TotalElements int
	TotalPages    int
	PageSize      int
}

var ErrNoMorePages = errors.New("the iterator has been consumed")
