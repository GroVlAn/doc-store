package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/GroVlAn/doc-store/internal/core"
	"github.com/GroVlAn/doc-store/internal/core/e"
	"github.com/go-chi/chi"
)

func (h *Handler) createDocument(w http.ResponseWriter, r *http.Request) {
	documentRequest := core.DocumentRequest{}

	metaValue := r.FormValue("meta")
	documentRequest.Meta = core.Document{}

	if err := json.Unmarshal([]byte(metaValue), &documentRequest.Meta); err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	if err := h.userService.VerifyAccessToken(documentRequest.Meta.Token); err != nil {
		h.sendErrorResponse(w, err)
		return
	}

	jsonValue := r.FormValue("json")

	var jsonData *interface{}
	if len(jsonValue) > 0 {
		if err := json.Unmarshal([]byte(jsonValue), &jsonData); err != nil {
			h.sendErrorResponse(w, err)
			return
		}

		documentRequest.Meta.Json = []byte(jsonValue)
	}

	file, header, err := r.FormFile("file")
	var b []byte
	if err == nil {
		defer file.Close()

		b, err = io.ReadAll(file)
		if err != nil {
			h.sendErrorResponse(w, err)

			return
		}

		documentRequest.Meta.Name = header.Filename
		documentRequest.Meta.Mime = http.DetectContentType(b)
	}

	if err := h.documentService.CreateDocument(documentRequest.Meta, b); err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}
	if header != nil {
		res.Response = struct {
			Json *interface{} `json:"json,omitempty"`
			File string       `json:"file"`
		}{
			Json: jsonData,
			File: header.Filename,
		}
	} else {
		res.Response = struct {
			Json *interface{} `json:"json,omitempty"`
		}{
			Json: jsonData,
		}
	}

	h.sendResponse(w, res, http.StatusCreated)
}

func (h *Handler) document(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "docID")

	body := r.Body
	defer h.closeRequestBody(body)

	var tokenBody struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(body).Decode(&tokenBody); err != nil {
		h.sendErrorResponse(w, &e.ErrInvalidToken{
			Msg: "invalid token",
			Err: fmt.Errorf("decoding token body: %w", err),
		})

		return
	}

	if err := h.userService.VerifyAccessToken(tokenBody.Token); err != nil {
		h.sendErrorResponse(w, err)
		return
	}

	document, file, err := h.documentService.Document(tokenBody.Token, docID)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	if document.Json != nil {
		h.serveJson(w, document.Json)
	} else {
		h.serveFile(w, r, file)
	}
}

func (h *Handler) documentsList(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer h.closeRequestBody(body)

	var docFilter core.DocumentFilter

	if err := json.NewDecoder(body).Decode(&docFilter); err != nil {
		h.sendErrorResponse(w, e.ErrEmptyBody)

		return
	}

	if err := h.userService.VerifyAccessToken(docFilter.Token); err != nil {
		h.sendErrorResponse(w, err)
		return
	}

	docList, err := h.documentService.DocumentsList(docFilter.Token, docFilter)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}
	res.Data = struct {
		Docs []core.Document `json:"docs,omitempty"`
	}{
		Docs: docList,
	}

	h.sendResponse(w, res, http.StatusCreated)
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "docID")

	body := r.Body
	defer h.closeRequestBody(body)

	var tokenBody struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(body).Decode(&tokenBody); err != nil {
		h.sendErrorResponse(w, &e.ErrInvalidToken{Msg: "invalid token", Err: err})

		return
	}

	if err := h.userService.VerifyAccessToken(tokenBody.Token); err != nil {
		h.sendErrorResponse(w, err)
		return
	}

	err := h.documentService.DeleteDocument(tokenBody.Token, docID)
	if err != nil {
		h.sendErrorResponse(w, err)

		return
	}

	res := core.Response{}
	res.Response = map[string]bool{docID: true}

	h.sendResponse(w, res, http.StatusCreated)
}

func (h *Handler) serveJson(w http.ResponseWriter, jsonData []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (h *Handler) serveFile(w http.ResponseWriter, r *http.Request, file string) {
	mineType := http.DetectContentType([]byte(file))

	w.Header().Set("Content-Type", mineType)
	http.ServeFile(w, r, file)
}
