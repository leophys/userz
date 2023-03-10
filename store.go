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
	List(ctx context.Context, filter *Filter, pageSize uint) (Iterator[[]*User], error)
	Page(ctx context.Context, filter *Filter, params *PageParams) ([]*User, error)
}

// UserData represents the data needed to create or alter a user.
type UserData struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	NickName  string `json:"nickname,omitempty"`
	Password  string `json:"password,omitempty"`
	Email     string `json:"email,omitempty"`
	Country   string `json:"country,omitempty"`
}

// PageParams conveys the information needed to specify a page for the Page
// method.
type PageParams struct {
	Size   uint
	Offset uint
	Order  Order
}

// Filter is a condition to be used to filter users. The backend type
// represents the output type a concrete implementation will produce
// as output of the evaluation of the filter.
type Filter struct {
	Id        string
	FirstName Condition[string]
	LastName  Condition[string]
	NickName  Condition[string]
	Email     Condition[string]
	Country   Condition[string]
	CreatedAt Condition[time.Time]
	UpdatedAt Condition[time.Time]
}

// Condition is the interface any backend will need to implement in order
// to translate the given condition in a valid expression for the backend
// at hand.
type Condition[T Conditionable] interface {
	// Evaluate translates the abstract Condition into a form that is
	// usable by the backend.
	Evaluate(field string) (any, error)
	// Hash returns a unique identified deterministically derived by the
	// values of the condition.
	Hash(field string) (string, error)
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
	TotalElements uint
	TotalPages    uint
	PageSize      uint
}

var ErrNoMorePages = errors.New("the iterator has been consumed")
