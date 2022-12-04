package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/leophys/userz"
	"github.com/leophys/userz/internal/httputils"
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

	id := strings.Trim(chi.URLParam(r, "id"), "\"")
	if id == "" {
		httputils.BadRequest(w, "Missing user id in request url")
		return
	}

	expiring, cancel := context.WithTimeout(ctx, defaultRemoveTimeout)
	defer cancel()

	user, err := h.store.Remove(expiring, id)
	if err != nil {
		logger.Err(err).Str("ID", id).Msg("Failure in removing the user")
		httputils.ServerError(w, "Failure in removing the user")
		return
	}

	logger.Info().Str("ID", user.Id).Msg("User removed")
	httputils.Ok(w, user)
}
