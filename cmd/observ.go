package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome")) //nolint:errcheck
	})
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
