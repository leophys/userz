package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/leophys/userz"
)

var (
	subsystem = "userz" // default subsystem, overwrite at compile time, if needed

	storeDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "store_request_duration_seconds",
	}, []string{"method"})
	storeFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "store_failures",
	}, []string{"method"})
)

func init() {
	prometheus.MustRegister(storeDuration)
	prometheus.MustRegister(storeFailures)
}

type MetricsStore struct {
	wrapped userz.Store
}

func NewMetricsStore(wrapped userz.Store) userz.Store {
	return &MetricsStore{
		wrapped: wrapped,
	}
}

func (s *MetricsStore) Add(ctx context.Context, user *userz.UserData) (*userz.User, error) {
	label := "Add"
	start := time.Now()

	res, err := s.wrapped.Add(ctx, user)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
	}
	storeDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())

	return res, err
}

func (s *MetricsStore) Update(ctx context.Context, id string, user *userz.UserData) (*userz.User, error) {
	label := "Update"
	start := time.Now()

	res, err := s.wrapped.Update(ctx, id, user)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
	}
	storeDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())

	return res, err
}

func (s *MetricsStore) Remove(ctx context.Context, id string) (*userz.User, error) {
	label := "Remove"
	start := time.Now()

	res, err := s.wrapped.Remove(ctx, id)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
	}
	storeDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())

	return res, err
}

func (s *MetricsStore) List(ctx context.Context, filter *userz.Filter, pageSize uint) (userz.Iterator[[]*userz.User], error) {
	label := "List"

	iterator, err := s.wrapped.List(ctx, filter, pageSize)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
		return nil, err
	}

	return &MetricsIterator{iterator}, nil
}

func (s *MetricsStore) Page(ctx context.Context, filter *userz.Filter, params *userz.PageParams) ([]*userz.User, error) {
	label := "Page"
	start := time.Now()

	res, err := s.wrapped.Page(ctx, filter, params)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
	}
	storeDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())

	return res, err
}

type MetricsIterator struct {
	wrapped userz.Iterator[[]*userz.User]
}

func (it *MetricsIterator) Len() userz.PaginationData {
	return it.wrapped.Len()
}

func (it *MetricsIterator) Next(ctx context.Context) ([]*userz.User, error) {
	label := "ListNext"
	start := time.Now()

	next, err := it.wrapped.Next(ctx)
	if err != nil {
		storeFailures.WithLabelValues(label).Inc()
	}
	storeDuration.WithLabelValues(label).Observe(time.Since(start).Seconds())

	return next, err
}
