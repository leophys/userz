package main

import (
	"container/ring"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/leophys/userz/internal/httputils"
	"github.com/leophys/userz/pkg/notifier"
)

const (
	defaultPort  = 8000
	defaultSize  = 10000
	defaultRoute = "/notifications"
)

var (
	envPort = "NOTIFIER_PORT"
	envSize = "NOTIFIER_BUFFER_SIZE"
)

var Provider notifier.Notifier = &polledNotifier{}

type notification struct {
	Event    notifier.NotificationEvent `json:"event"`
	Metadata map[string]string          `json:"metadata,omitempty"`
}

type polledNotifier struct {
	notify chan *notification
	buf    *ring.Ring

	mu sync.Mutex
}

func (n *polledNotifier) Init(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	port := defaultPort
	size := defaultSize

	if envPortStr := os.Getenv(envPort); envPortStr != "" {
		envPort, err := strconv.Atoi(envPortStr)
		if err != nil {
			return err
		}
		port = envPort
	}

	if envSizeStr := os.Getenv(envSize); envSizeStr != "" {
		envSize, err := strconv.Atoi(envSizeStr)
		if err != nil {
			return err
		}
		size = envSize
	}

	router := chi.NewRouter()
	if logger != nil {
		router.Use(httputils.LoggerMiddleware(*logger))
	}

	router.Get(defaultRoute, func(w http.ResponseWriter, r *http.Request) {
		n.mu.Lock()
		defer n.mu.Unlock()

		resp := []*notification{}

		logger := zerolog.Ctx(r.Context()).
			With().
			Str("Handler", "Notifier").
			Logger()

		n.buf.Do(func(item any) {
			if item != nil {
				resp = append(resp, item.(*notification))
			}
		})

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Err(err).Msg("Failed to serialize notifications")
			httputils.ServerError(w, "Failed to serialize notifications")
			return
		}

		n.buf = ring.New(size)

		logger.Info().Msg("Notification buffer flushed by polling")
	})

	addr := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger.Info().
		Int("size", size).
		Msgf("Serving polled notifier on '%s'", addr)

	n.notify = make(chan *notification, size)
	n.buf = ring.New(size)

	go n.listen(ctx)

	go func() {
		if err := http.Serve(listener, router); err != nil {
			logger.Err(err).Msg("Failed to serve polled notifier handler")
		}
	}()

	return nil
}

func (n *polledNotifier) Notify(ctx context.Context, event notifier.NotificationEvent, metadata map[string]string) error {
	n.notify <- &notification{
		Event:    event,
		Metadata: metadata,
	}

	return nil
}

func (n *polledNotifier) listen(ctx context.Context) {
	logger := zerolog.Ctx(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-n.notify:
			logger.Debug().
				Str("notificationType", notification.Event.String()).
				Msg("Notification received")

			n.mu.Lock()
			n.buf.Value = notification
			n.buf = n.buf.Next()
			n.mu.Unlock()
		}
	}
}
