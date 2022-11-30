package httpapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leophys/userz"
)

func TestPageHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	users := []*userz.User{
		{Id: "1"},
		{Id: "2"},
		{Id: "3"},
		{Id: "4"},
		{Id: "5"},
		{Id: "6"},
	}
	store := &mockStore{data: users}
	h := &PageHandler{store}
	router := chi.NewRouter()
	router.Get("/", h.ServeHTTP)

    // get the first page
	req := httptest.NewRequest(http.MethodGet, localhost+"?pageSize=3&offset=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()
	if !assert.Equal(http.StatusOK, resp.StatusCode) {
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(err)
		t.Log(string(body))
	}

	var result []*userz.User
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(users[:3], result)

    // get the second page
	req = httptest.NewRequest(http.MethodGet, localhost+"?pageSize=3&offset=3", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp = w.Result()
	if !assert.Equal(http.StatusOK, resp.StatusCode) {
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(err)
		t.Log(string(body))
	}

    result = nil
	json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(users[3:], result)

    // try to get to a third page, get 404
	req = httptest.NewRequest(http.MethodGet, localhost+"?pageSize=3&offset=6", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp = w.Result()
    require.Equal(http.StatusNotFound, resp.StatusCode)

	assert.Equal(3, store.paged)
}
