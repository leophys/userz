package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/leophys/userz"
)

func TestRemoveHandler(t *testing.T) {
	assert := assert.New(t)

	user := &userz.User{
		Id: "1",
	}
	store := &mockStore{data: []*userz.User{user}}
	h := &RemoveHandler{store}
	router := chi.NewRouter()
	router.Delete("/{id}", h.ServeHTTP)

	// Malformed request
	req := httptest.NewRequest(http.MethodDelete, localhost, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(http.StatusNotFound, resp.StatusCode)

	// Correct request
	req = httptest.NewRequest(http.MethodDelete, localhost+"1", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)

	var result userz.User
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(user, &result)

	assert.Equal(1, store.removed)
}
