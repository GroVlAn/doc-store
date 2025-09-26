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
	documentPath = "/docs"
)

type userService interface {
	Register(user core.User) error
	Auth(user core.User) (string, error)
	VerifyAccessToken(token string) error
	Logout(token string) error
}

type documentService interface {
	CreateDocument(document core.Document, file []byte) error
	Document(token string, documentID string) (core.Document, string, error)
	DocumentsList(token string, filter core.DocumentFilter) ([]core.Document, error)
	DeleteDocument(token string, documentID string) error
}

type Handler struct {
	l               zerolog.Logger
	userService     userService
	documentService documentService
}

func New(l zerolog.Logger, userService userService, documentService documentService) *Handler {
	return &Handler{
		l:               l,
		userService:     userService,
		documentService: documentService,
	}
}

func (h *Handler) Handler() *chi.Mux {
	r := chi.NewRouter()

	h.useMiddleware(r)

	r.Route(addrPath, func(r chi.Router) {
		r.Post(registerPath, h.register)
		r.Get(authPath, h.auth)
		r.Delete(authPath+"/{token}", h.logout)

		r.Post(documentPath, h.createDocument)
		r.Get(documentPath, h.documentsList)

		r.Route(documentPath+"/{docID}", func(r chi.Router) {
			r.Get("/", h.document)
			r.Delete("/", h.deleteDocument)
		})
	})

	return r
}

func (h *Handler) sendErrorResponse(w http.ResponseWriter, err error) {
	res := core.Response{}

	status, msg := h.handleError(err)

	res.Error = &core.ErrorResponse{}
	res.Error.Code = status
	res.Error.Text = msg

	h.sendResponse(w, res, status)
}

func (h *Handler) handleError(err error) (int, string) {
	var errFind *e.ErrFind
	var errInsert *e.ErrInsert
	var errDelete *e.ErrDelete
	var errInvalidToken *e.ErrInvalidToken

	switch {
	case errors.As(err, &errInvalidToken):
		h.l.Error().Err(errInvalidToken.Unwrap()).Msg("failed verify token")

		return http.StatusUnauthorized, err.Error()
	case errors.As(err, &errFind):
		h.l.Error().Err(errFind.Unwrap()).Msg("failed to find")

		return http.StatusNotFound, errFind.Error()
	case errors.As(err, &errInsert):
		h.l.Error().Err(errInsert.Unwrap()).Msg("failed to create")

		return http.StatusInternalServerError, errFind.Error()
	case errors.As(err, &errDelete):
		h.l.Error().Err(errDelete.Unwrap()).Msg("failed to delete")

		return http.StatusInternalServerError, errDelete.Error()
	case errors.Is(err, e.ErrInvalidLogin):
		h.l.Error().Err(err).Msg("failed to verify login")

		return http.StatusBadRequest, err.Error()
	case errors.Is(err, e.ErrInvalidPassword):
		h.l.Error().Err(err).Msg("failed to verify password")

		return http.StatusBadRequest, err.Error()
	case errors.Is(err, e.ErrUserAlreadyExist):
		h.l.Error().Err(err).Msg("failed to create new user")

		return http.StatusBadRequest, err.Error()
	case errors.Is(err, e.ErrUserNotFound):
		h.l.Error().Err(err).Msg("failed to find exited user")

		return http.StatusNotFound, err.Error()
	case errors.Is(err, e.ErrNoDocuments):
		h.l.Error().Err(err).Msg("document not found")

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
