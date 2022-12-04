package notifying

import (
	"context"

	"github.com/leophys/userz"
	"github.com/leophys/userz/pkg/notifier"
)

var _ userz.Store = &NotifyingStore{}

type NotifyingStore struct {
	wrapped  userz.Store
	provider notifier.Notifier
}

func NewNotifyingStore(wrapped userz.Store, provider notifier.Notifier) userz.Store {
	return &NotifyingStore{
		wrapped:  wrapped,
		provider: provider,
	}
}

func (s *NotifyingStore) Add(ctx context.Context, user *userz.UserData) (*userz.User, error) {
	res, err := s.wrapped.Add(ctx, user)
	if err == nil {
		if err := s.provider.Notify(ctx, notifier.NotifyAccountCreated, map[string]string{
			"id": res.Id,
		}); err != nil {
			return nil, err
		}
	}

	return res, err
}

func (s *NotifyingStore) Update(ctx context.Context, id string, user *userz.UserData) (*userz.User, error) {
	res, err := s.wrapped.Update(ctx, id, user)
	if err == nil {
		if err := s.provider.Notify(ctx, notifier.NotifyAccountUpdated, map[string]string{
			"id": id,
		}); err != nil {
			return nil, err
		}
	}

	return res, err
}

func (s *NotifyingStore) Remove(ctx context.Context, id string) (*userz.User, error) {
	res, err := s.wrapped.Remove(ctx, id)
	if err == nil {
		if err := s.provider.Notify(ctx, notifier.NotifyAccountRemoved, map[string]string{
			"id": id,
		}); err != nil {
			return nil, err
		}
	}

	return res, err
}

func (s *NotifyingStore) List(ctx context.Context, filter *userz.Filter, pageSize uint) (userz.Iterator[[]*userz.User], error) {
	return s.wrapped.List(ctx, filter, pageSize)
}

func (s *NotifyingStore) Page(ctx context.Context, filter *userz.Filter, params *userz.PageParams) ([]*userz.User, error) {
	return s.wrapped.Page(ctx, filter, params)
}
