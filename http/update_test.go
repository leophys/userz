package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leophys/userz"
)

func TestUpdateHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	userData := userz.UserData{
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "jd",
		Password:  "passw0rd",
		Email:     "jd@morgue.com",
		Country:   "US",
	}

	b := bytes.NewBuffer(nil)
	err := json.NewEncoder(b).Encode(&userData)
	require.NoError(err)

	user := &userz.User{
		Id: "1",
	}
	store := &mockStore{data: []*userz.User{user}}
	h := &UpdateHandler{store}
	router := chi.NewRouter()
	router.Post("/{id}", h.ServeHTTP)

	// Malformed request
	req := httptest.NewRequest(http.MethodPost, localhost, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(http.StatusNotFound, resp.StatusCode)

	// Correct request
	req = httptest.NewRequest(http.MethodPost, localhost+"1", b)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)

	var result userz.User
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(user, &result)

	assert.Equal(1, store.updated)
}
