package api

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

// DocumentResponse represents a document key and its accessible URL.
type DocumentResponse struct {
	Key string `json:"key" example:"123e4567-e89b-12d3-a456-426614174000.pdf"`
	URL string `json:"url" example:"http://localhost:9000/documents/123e4567-e89b-12d3-a456-426614174000.pdf?X-Amz-Algorithm=..."`
}

// DocumentsResponse is the response payload for listing documents.
type DocumentsResponse struct {
	Documents []DocumentResponse `json:"documents" validate:"required"`
}

// ListDocuments godoc
// @Summary List documents
// @Description Retrieves a list of documents; supports optional substring search
// @Tags documents
// @Produce json
// @Param query query string false "Search query"
// @Success 200 {object} DocumentsResponse
// @Failure 500 {object} Error "Internal server error"
// @Router /documents [get]
func (a *API) ListDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	query := r.URL.Query().Get("query")
	keys, err := a.s3.ListObjects(ctx, "", true)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}

	var docs []DocumentResponse
	for _, key := range keys {
		if query != "" && !strings.Contains(key, query) {
			continue
		}
		url, err := a.s3.PresignGet(ctx, key, 15*time.Minute)
		if err != nil {
			rec.Add(events.Error, err)
			writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
			return
		}
		docs = append(docs, DocumentResponse{
			Key: key,
			URL: url.String(),
		})
	}

	a.writeJSON(ctx, w, DocumentsResponse{Documents: docs}, http.StatusOK)
}

// UploadDocument godoc
// @Summary Upload document
// @Description Uploads a new document to storage
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file"
// @Success 201 {object} DocumentResponse
// @Failure 400 {object} InvalidRequestError "Invalid request format"
// @Failure 500 {object} Error "Internal server error"
// @Router /documents [post]
func (a *API) UploadDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(ctx, w, ErrInvalidRequest.WithDetails("file form field required").WithStatus(http.StatusBadRequest))
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	id, _ := uuid.NewV4()
	key := id.String() + ext

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	size := header.Size
	url, err := a.s3.Store(ctx, key, file.(io.Reader), size, contentType, 24*time.Hour)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}

	a.writeJSON(ctx, w, DocumentResponse{Key: key, URL: url.String()}, http.StatusCreated)
}

// DeleteDocument godoc
// @Summary Delete document
// @Description Deletes a document by its key
// @Tags documents
// @Param key query string true "Document key"
// @Success 204 "No content"
// @Failure 400 {object} InvalidRequestError "Missing document key"
// @Failure 500 {object} Error "Internal server error"
// @Router /documents [delete]
func (a *API) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	key := r.URL.Query().Get("key")
	if key == "" {
		writeError(ctx, w, ErrInvalidRequest.WithDetails("key query parameter required").WithStatus(http.StatusBadRequest))
		return
	}

	if err := a.s3.DeleteObject(ctx, key); err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDocument godoc
// @Summary Get document URL
// @Description Retrieves a presigned URL for a document by key
// @Tags documents
// @Produce json
// @Param id path string true "Document key"
// @Success 200 {object} DocumentResponse
// @Failure 400 {object} InvalidRequestError "Invalid document key"
// @Failure 500 {object} Error "Internal server error"
// @Router /documents/{id} [get]
func (a *API) GetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx)

	key := chi.URLParam(r, "id")
	if key == "" {
		writeError(ctx, w, ErrInvalidRequest.WithDetails("id path parameter required").WithStatus(http.StatusBadRequest))
		return
	}

	url, err := a.s3.PresignGet(ctx, key, 15*time.Minute)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}

	a.writeJSON(ctx, w, DocumentResponse{Key: key, URL: url.String()}, http.StatusOK)
}
