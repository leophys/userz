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

func TestAddHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	password, _ := userz.NewPassword("passw0rd")

	userData := userz.UserData{
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "jd",
		Password:  password,
		Email:     "jd@morgue.com",
		Country:   "US",
	}

	b := bytes.NewBuffer(nil)
	err := json.NewEncoder(b).Encode(&userData)
	require.NoError(err)

	req := httptest.NewRequest(http.MethodPut, localhost, b)
	w := httptest.NewRecorder()

	user := &userz.User{
		Id: "1",
	}
	store := &mockStore{data: []*userz.User{user}}
	h := &AddHandler{store}
	router := chi.NewRouter()
	router.Put("/", h.ServeHTTP)

	router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(http.StatusOK, resp.StatusCode)

	var result string
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(user.Id, result)
	assert.Equal(1, store.added)
}
