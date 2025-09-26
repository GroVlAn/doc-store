package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/go-chi/chi"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer h.closeRequestBody(body)

	var (
		user core.User
	)

	err := json.NewDecoder(body).Decode(&user)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	err = h.userService.Register(user)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}
	res.Response = struct {
		Login string `json:"login"`
	}{
		Login: user.Login,
	}

	h.sendResponse(w, res, http.StatusCreated)
}

func (h *Handler) auth(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer h.closeRequestBody(body)

	var user core.User

	err := json.NewDecoder(body).Decode(&user)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	token, err := h.userService.Auth(user)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}

	res.Response = struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	h.sendResponse(w, res, http.StatusOK)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := h.userService.Logout(token)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}
	res.Response = map[string]bool{token: true}

	h.sendResponse(w, res, http.StatusOK)
}
