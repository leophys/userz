package httpapi

import (
	"context"

	"github.com/leophys/userz"
)

const localhost = "http://localhost/"

var _ userz.Store = &mockStore{}

type mockStore struct {
	added   int
	removed int
	updated int
	paged   int

	data []*userz.User
}

func (s *mockStore) Add(ctx context.Context, user *userz.UserData) (*userz.User, error) {
	s.added++
	u := s.data[0]
	s.data = s.data[1:]
	return u, nil
}

func (s *mockStore) Update(ctx context.Context, id string, user *userz.UserData) (*userz.User, error) {
	s.updated++
	u := s.data[0]
	s.data = s.data[1:]
	return u, nil
}

func (s *mockStore) Remove(ctx context.Context, id string) (*userz.User, error) {
	s.removed++
	u := s.data[0]
	s.data = s.data[1:]
	return u, nil
}

func (s *mockStore) List(ctx context.Context, filter *userz.Filter, pageSize uint) (userz.Iterator[[]*userz.User], error) {
	return nil, nil
}

func (s *mockStore) Page(ctx context.Context, filter *userz.Filter, params *userz.PageParams) ([]*userz.User, error) {
	s.paged++
	if uint(len(s.data)) < params.Size {
		return nil, nil
	}
	users := s.data[0:params.Size]
	s.data = s.data[params.Size:]
	return users, nil
}
