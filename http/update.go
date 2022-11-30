package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/leophys/userz"
)

const (
	defaultUpdateTimeout = 30 * time.Second
)

var _ http.Handler = &UpdateHandler{}

type UpdateHandler struct {
	store userz.Store
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).
		With().
		Str("Handler", "UpdateHandler").
		Logger()

	id := chi.URLParam(r, "id")
	if id == "" {
		badRequest(w, "Missing user id in request url")
		return
	}

	var userData userz.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		logger.Err(err).Msg("Failure in decoding request body")
		badRequest(w, "Malformed request body")
		return
	}

	expiring, cancel := context.WithTimeout(ctx, defaultUpdateTimeout)
	defer cancel()

	user, err := h.store.Update(expiring, id, &userData)
	if err != nil {
		logger.Err(err).Str("ID", id).Msg("Failure in updating the user")
		serverError(w, "Failure in updating the user")
		return
	}
	if user == nil {
		logger.Warn().Err(err).Str("ID", id).Msg("Missing user")
		notFound(w, "No user found")
		return
	}

	logger.Info().Str("ID", user.Id).Msg("User updated")
	ok(w, user)
}
