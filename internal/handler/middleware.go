package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (h *HTTPHandler) useMiddleware(r *chi.Mux) {
	r.Use(h.cors)
}

func (h *HTTPHandler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}
