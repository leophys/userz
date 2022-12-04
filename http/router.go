package httpapi

import (
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/leophys/userz"
	"github.com/leophys/userz/internal/httputils"
)

func New(baseRoute string, store userz.Store, logger *zerolog.Logger) chi.Router {
	router := chi.NewRouter()

	if logger != nil {
		router.Use(httputils.LoggerMiddleware(*logger))
	}

	base := strings.TrimRight(baseRoute, "/")

	page := &PageHandler{store}
	router.Get(base, page.ServeHTTP)

	add := &AddHandler{store}
	router.Put(base, add.ServeHTTP)

	update := &UpdateHandler{store}
	router.Post(base+"/{id}", update.ServeHTTP)

	remove := &RemoveHandler{store}
	router.Delete(base+"/{id}", remove.ServeHTTP)

	return router
}
