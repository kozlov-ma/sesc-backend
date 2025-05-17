package api

import (
	"fmt"
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

// getScheme returns the scheme (http or https) from the request
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

// DocumentResponse represents a document key and download URL.
type DocumentResponse struct {
	Key string `json:"key" example:"123e4567-e89b-12d3-a456-426614174000.pdf"`
	URL string `json:"url" example:"http://api-server:8080/documents/123e4567-e89b-12d3-a456-426614174000.pdf"`
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
		// Build direct API endpoint URL instead of using presigned URLs
		downloadURL := fmt.Sprintf("%s://%s/documents/%s",
			getScheme(r), r.Host, key)
		docs = append(docs, DocumentResponse{
			Key: key,
			URL: downloadURL,
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
	err = a.s3.Store(ctx, key, file.(io.Reader), size, contentType, 24*time.Hour)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}

	// Build direct API endpoint URL instead of using presigned URLs
	downloadURL := fmt.Sprintf("%s://%s/documents/%s",
		getScheme(r), r.Host, key)
	a.writeJSON(ctx, w, DocumentResponse{Key: key, URL: downloadURL}, http.StatusCreated)
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
// @Summary Download document
// @Description Streams document content directly to client
// @Tags documents
// @Produce */*
// @Param id path string true "Document key"
// @Success 200 {file} binary "Document content"
// @Failure 400 {object} InvalidRequestError "Invalid document key"
// @Failure 404 {object} Error "Document not found"
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

	obj, info, err := a.s3.GetObject(ctx, key)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrServerError.WithDetails(err.Error()).WithStatus(http.StatusInternalServerError))
		return
	}
	defer obj.Close()

	filename := key
	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))

	if _, err := io.Copy(w, obj); err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, ErrInvalidRequest.WithDetails(err.Error()))
	}
}
