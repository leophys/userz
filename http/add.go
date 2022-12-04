package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/leophys/userz"
	"github.com/leophys/userz/internal/httputils"
)

const (
	defaultAddTimeout = 30 * time.Second
)

var _ http.Handler = &AddHandler{}

type AddHandler struct {
	store userz.Store
}

func (h *AddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).
		With().
		Str("Handler", "AddHandler").
		Logger()

	logger.Debug().Msg("Adding user")

	var userData userz.UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		logger.Err(err).Msg("Failure in decoding request body")
		httputils.BadRequest(w, "Malformed request body")
		return
	}

	expiring, cancel := context.WithTimeout(ctx, defaultAddTimeout)
	defer cancel()

	newUser, err := h.store.Add(expiring, &userData)
	if err != nil {
		logger.Err(err).Msg("Failure in adding the user")
		httputils.ServerError(w, "Failure in adding the user in the store")
		return
	}

	logger.Info().Str("ID", newUser.Id).Msg("New user added")
	httputils.Ok(w, newUser.Id)
}
