package handler

import (
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

type HTTPHandler struct {
	l zerolog.Logger
}

func New(l zerolog.Logger) *HTTPHandler {
	return &HTTPHandler{
		l: l,
	}
}

func (h *HTTPHandler) Handler() *chi.Mux {
	r := chi.NewRouter()

	h.useMiddleware(r)

	return r
}
