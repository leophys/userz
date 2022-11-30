package httpapi

import (
	"encoding/json"
	"net/http"
)

func ok(w http.ResponseWriter, resp any) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func notFound(w http.ResponseWriter, msg any) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func badRequest(w http.ResponseWriter, errMsg string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"error": errMsg,
	})
}

func serverError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"reason": msg,
	})
}
