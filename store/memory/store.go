package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/leophys/userz"
)

var _ userz.Store = &MemoryStore{}

type MemoryStore struct {
	data map[string]*userz.User
	mu   sync.Mutex
}

func NewMemoryStore() userz.Store {
	return &MemoryStore{
		data: make(map[string]*userz.User),
	}
}

func (s *MemoryStore) Add(ctx context.Context, user *userz.UserData) (*userz.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()

	password, err := userz.NewPassword(user.Password)
	if err != nil {
		return nil, err
	}

	newUser := &userz.User{
		Id:        id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		NickName:  user.NickName,
		Email:     user.Email,
		Password:  password,
		Country:   user.Country,
		CreatedAt: time.Now(),
	}
	s.data[id] = newUser

	return newUser, nil
}

func (s *MemoryStore) Update(ctx context.Context, id string, user *userz.UserData) (*userz.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	curUser, ok := s.data[id]
	if !ok {
		return nil, nil
	}

	if user.FirstName != "" {
		curUser.FirstName = user.FirstName
	}

	if user.LastName != "" {
		curUser.LastName = user.LastName
	}

	if user.Country != "" {
		curUser.Country = user.Country
	}

	if user.Email != "" {
		curUser.Email = user.Email
	}

	if user.Password != "" {
		newPassword, err := userz.NewPassword(user.Password)
		if err != nil {
			return nil, err
		}

		curUser.Password = newPassword
	}

	curUser.UpdatedAt = time.Now()

	s.data[id] = curUser

	return curUser, nil
}

func (s *MemoryStore) Remove(ctx context.Context, id string) (*userz.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.data[id]
	if !ok {
		return nil, nil
	}

	delete(s.data, id)

	return user, nil
}

func (s *MemoryStore) List(ctx context.Context, filter *userz.Filter, pageSize uint) (userz.Iterator[[]*userz.User], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// NOTE: this store ignores the filter
	var paginated [][]*userz.User
	elems := len(s.data)

	var page []*userz.User
	var counter uint
	for _, user := range s.data {
		counter++
		page = append(page, user)
		if counter == pageSize {
			paginated = append(paginated, page)
			page = nil
			counter = 0
		}
	}

	if counter != 0 {
		paginated = append(paginated, page)
	}

	iterator := &MemoryIterator{
		info: userz.PaginationData{
			TotalElements: uint(elems),
			TotalPages:    uint(len(paginated)),
			PageSize:      pageSize,
		},
		data: paginated,
	}

	return iterator, nil
}

func (s *MemoryStore) Page(ctx context.Context, filter *userz.Filter, params *userz.PageParams) ([]*userz.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var page []*userz.User

	var counter uint
	for _, user := range s.data {
		if counter >= params.Offset {
			page = append(page, user)
		}

		counter++

		if counter == params.Offset+params.Size {
			break
		}
	}

	return page, nil
}

type MemoryIterator struct {
	info userz.PaginationData
	data [][]*userz.User
}

func (i *MemoryIterator) Len() userz.PaginationData {
	return i.info
}

func (i *MemoryIterator) Next(ctx context.Context) ([]*userz.User, error) {
	if len(i.data) == 0 {
		return nil, userz.ErrNoMorePages
	}

	resp := i.data[0]
	i.data = i.data[1:]

	return resp, nil
}
