package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/GroVlAn/doc-store/internal/core/e"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

const (
	addrPath     = "/api"
	registerPath = "/register"
	authPath     = "/auth"
)

type userService interface {
	Register(user core.User) error
	Auth(user core.User) (string, error)
	VerifyAccessToken(token string) error
}

type Handler struct {
	l           zerolog.Logger
	userService userService
}

func New(l zerolog.Logger, userService userService) *Handler {
	return &Handler{
		l:           l,
		userService: userService,
	}
}

func (h *Handler) Handler() *chi.Mux {
	r := chi.NewRouter()

	h.useMiddleware(r)

	r.Route(addrPath, func(r chi.Router) {
		r.Post(registerPath, h.register)
		r.Get(authPath, h.auth)
	})

	return r
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer h.closeRequestBody(body)

	var (
		user core.User
		res  core.Response = core.Response{}
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

func (h *Handler) sendErrorResponse(w http.ResponseWriter, err error) {
	res := core.Response{}

	status, msg := h.handleError(err)

	res.Error.Code = status
	res.Error.Text = msg

	h.sendResponse(w, res, status)
}

func (h *Handler) handleError(err error) (int, string) {
	var errFind *e.ErrFind
	var errInsert *e.ErrInsert

	switch {
	case errors.As(err, &errFind):
		h.l.Error().Err(err.(*e.ErrFind).Unwrap()).Msg("failed to find user")

		return http.StatusNotFound, err.(*e.ErrFind).Error()
	case errors.As(err, &errInsert):
		h.l.Error().Err(err.(*e.ErrInsert).Unwrap()).Msg("failed to create user")

		return http.StatusInternalServerError, err.(*e.ErrInsert).Error()
	case errors.Is(err, e.ErrInvalidPassword):
		h.l.Error().Err(err).Msg("failed to verify password")

		return http.StatusBadRequest, err.Error()
	case errors.Is(err, e.ErrUserAlreadyExist):
		h.l.Error().Err(err).Msg("failed to create new user")

		return http.StatusBadRequest, err.Error()
	case errors.Is(err, e.ErrUserNotFound):
		h.l.Error().Err(err).Msg("failed to find exited user")

		return http.StatusNotFound, err.Error()
	default:
		h.l.Error().Err(err).Msg("internal error")

		return http.StatusInternalServerError, "internal error"
	}
}

func (h *Handler) sendResponse(w http.ResponseWriter, res core.Response, status int) {
	b, err := json.Marshal(res)
	if err != nil {
		h.l.Error().Err(err).Msg("failed marshal response")
	}

	w.WriteHeader(status)

	_, err = w.Write(b)
	if err != nil {
		h.l.Error().Err(err).Msg("failed write response")
	}
}

func (h *Handler) closeRequestBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		h.l.Error().Err(err).Msg("failed to close request body")
	}
}
