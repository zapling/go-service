package handler

import (
	"net/http"

	"github.com/zapling/go-service/internal/business"
)

func Foo(bc *business.Client) http.HandlerFunc {

	type response struct {
		Bar string `json:"bar"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		WriteJson(w, r, http.StatusOK, response{Bar: "baz"})
	}
}
