package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/leophys/userz"
)

const (
	defaultPageTimeout = 30 * time.Second
)

var _ http.Handler = &PageHandler{}

type PageHandler struct {
	store userz.Store
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).
		With().
		Str("Handler", "PageHandler").
		Logger()

	params := parsePageParams(w, r, &logger)
	if params == nil {
		return
	}

	filter, success := parseFilter(w, r, &logger)
	if !success {
		return
	}

	logger.Debug().
		Interface("page", params).
		Interface("filter", filter).
		Msg("Current scope of page request")

	expiring, cancel := context.WithTimeout(ctx, defaultPageTimeout)
	defer cancel()

	users, err := h.store.Page(expiring, filter, params)
	if err != nil {
		logger.Err(err).Msg("Failure in retrieving the users")
		serverError(w, "Failure in retrieving the users")
		return
	}

	if users == nil {
		logger.Debug().Msg("No users found")
		notFound(w, "No more pages")
		return
	}

	var ids []string
	for _, u := range users {
		ids = append(ids, u.Id)
	}
	logger.Info().Strs("ID", ids).Msg("Users retrieved")
	ok(w, users)
}

func parsePageParams(w http.ResponseWriter, r *http.Request, logger *zerolog.Logger) *userz.PageParams {
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		logger.Debug().Msg("Missing pageSize in request url")
		badRequest(w, "Missing pageSize in request url")
		return nil
	}
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)
	if err != nil {
		logger.Debug().Msg("pageSize must be a non negative integer")
		badRequest(w, "pageSize must be a non negative integer")
		return nil
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		logger.Debug().Msg("Missing offset in request url")
		badRequest(w, "Missing offset in request url")
		return nil
	}
	offset, err := strconv.ParseUint(offsetStr, 10, 32)
	if err != nil {
		logger.Debug().Msg("offset must be a non negative integer")
		badRequest(w, "offset must be a non negative integer")
		return nil
	}

	ord, err := userz.ParseOrder(r.URL.Query().Get("order_by"), r.URL.Query().Get("order_dir"))
	if err != nil {
		logger.Info().Err(err).Msg("Unacceptable order_by")
		badRequest(w, "unacceptable order_by")
		return nil
	}

	return &userz.PageParams{
		Size:   uint(pageSize),
		Offset: uint(offset),
		Order:  ord,
	}
}

func parseFilter(w http.ResponseWriter, r *http.Request, logger *zerolog.Logger) (*userz.Filter, bool) {
	params := make(map[string]string)

	if v := r.URL.Query().Get("first_name"); v != "" {
		params["first_name"] = v
	}

	if v := r.URL.Query().Get("last_name"); v != "" {
		params["last_name"] = v
	}

	if v := r.URL.Query().Get("nickname"); v != "" {
		params["nickname"] = v
	}

	if v := r.URL.Query().Get("email"); v != "" {
		params["email"] = v
	}

	if v := r.URL.Query().Get("country"); v != "" {
		params["country"] = v
	}

	if v := r.URL.Query().Get("created_at"); v != "" {
		params["created_at"] = v
	}

	if v := r.URL.Query().Get("updated_at"); v != "" {
		params["updated_at"] = v
	}

	filter, err := userz.ParseFilter(params)
	if err != nil {
		logger.Info().Err(err).Msg("Malformed filter")
		badRequest(w, "Malformed filter")
		return nil, false
	}

	return filter, true
}
