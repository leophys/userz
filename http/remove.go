package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/leophys/userz"
)

const (
	defaultRemoveTimeout = 30 * time.Second
)

var _ http.Handler = &RemoveHandler{}

type RemoveHandler struct {
	store userz.Store
}

func (h *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).
		With().
		Str("Handler", "RemoveHandler").
		Logger()

	id := chi.URLParam(r, "id")
	if id == "" {
		badRequest(w, "Missing user id in request url")
		return
	}

	expiring, cancel := context.WithTimeout(ctx, defaultRemoveTimeout)
	defer cancel()

	user, err := h.store.Remove(expiring, id)
	if err != nil {
		logger.Err(err).Str("ID", id).Msg("Failure in removing the user")
		serverError(w, "Failure in removing the user")
		return
	}

	logger.Info().Str("ID", user.Id).Msg("User removed")
	ok(w, user)
}
