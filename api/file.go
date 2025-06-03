package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// FileResponse is the API response for a file
type FileResponse struct {
	ID          string  `json:"id"`
	OwnerID     *string `json:"ownerId,omitempty"`
	FileName    string  `json:"fileName"`
	FileSize    int     `json:"fileSize"`
	DownloadURL string  `json:"downloadUrl"`
}

// FileListResponse is the API response for a list of files
type FileListResponse struct {
	Items      []FileResponse `json:"items"`
	TotalCount int            `json:"totalCount"`
}

// FileAccessMiddleware checks if the user has access to the requested file
func (a *API) FileAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx).Sub("file_access_check")

		// Extract file ID from URL
		fileID, err := uuid.FromString(chi.URLParam(r, "id"))
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		// Get identity from context
		identity, identityOk := GetIdentityFromContext(ctx)
		if !identityOk {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		// Get file details to check ownership
		file, err := a.file.ByID(ctx, fileID)
		if err != nil {
			if errors.Is(err, sesc.ErrFileNotFound) {
				a.writeJSON(ctx, w, respond.WithError(ctx, err))
				return
			}
			rec.Add(events.Error, err)
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		// Check access permissions:
		// 1. Common files can be accessed by anyone
		isCommonFile := file.OwnerID == nil
		// 2. Admin users can access any file
		isAdmin := identity.Role == iam.Role("admin")
		// 3. File owners can access their own files
		isOwner := file.OwnerID != nil && identity.ID == *file.OwnerID

		if isCommonFile || isAdmin || isOwner {
			// User has access, continue to the actual handler
			next.ServeHTTP(w, r)
			return
		}

		// Access denied
		rec.Add("access_denied", true)
		rec.Add("reason", "user is not owner, not admin, and file is not common")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
	})
}

func (a *API) FileEditAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx).Sub("file_access_check")

		// Extract file ID from URL
		fileID, err := uuid.FromString(chi.URLParam(r, "id"))
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		// Get identity from context
		identity, identityOk := GetIdentityFromContext(ctx)
		if !identityOk {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		// Get file details to check ownership
		file, err := a.file.ByID(ctx, fileID)
		if err != nil {
			if errors.Is(err, sesc.ErrFileNotFound) {
				a.writeJSON(ctx, w, respond.WithError(ctx, err))
				return
			}
			rec.Add(events.Error, err)
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}

		isAdmin := identity.Role == iam.Role("admin")
		isOwner := file.OwnerID != nil && identity.ID == *file.OwnerID

		if isAdmin || isOwner {
			// User has access, continue to the actual handler
			next.ServeHTTP(w, r)
			return
		}

		// Access denied
		rec.Add("access_denied", true)
		rec.Add("reason", "user is not owner, not admin, and file is not common")
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
	})
}

// convertFile converts a sesc.File to a FileResponse
func convertFile(f sesc.File) FileResponse {
	var ownerID *string
	if f.OwnerID != nil {
		id := f.OwnerID.String()
		ownerID = &id
	}

	return FileResponse{
		ID:          f.ID.String(),
		OwnerID:     ownerID,
		FileName:    f.Name,
		FileSize:    f.Size,
		DownloadURL: f.URL,
	}
}

// SearchFiles returns a list of files matching the search criteria
// @Summary Search files
// @Description Returns a list of files based on search criteria
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param name query string false "File name to search for"
// @Param owner_id query string false "Owner ID to filter by"
// @Param common query bool false "If true, return only common files"
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(50)
// @Success 200 {object} FileListResponse
// @Failure 400 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /files [get]
// @Security BearerAuth
func (a *API) SearchFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/search_files")

	query := r.URL.Query()

	opts := sesc.FileSearchOptions{
		Name: query.Get("name"),
	}

	if ownerIDStr := query.Get("owner_id"); ownerIDStr != "" {
		if ide, ok := GetIdentityFromContext(ctx); ok && ownerIDStr == "me" {
			opts.OwnerID = &ide.ID
		} else {
			ownerID, err := uuid.FromString(ownerIDStr)
			if err != nil {
				a.writeJSON(ctx, w, respond.WithError(ctx, err))
				return
			}
			opts.OwnerID = &ownerID
		}
	}

	if commonStr := query.Get("common"); commonStr != "" {
		common, err := strconv.ParseBool(commonStr)
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		opts.Common = common
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		opts.Offset = offset
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		opts.Limit = limit
	}

	files, totalCount, err := a.file.Search(ctx, opts)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := FileListResponse{
		Items:      make([]FileResponse, len(files)),
		TotalCount: totalCount,
	}

	for i, file := range files {
		response.Items[i] = convertFile(file)
	}

	a.writeJSON(ctx, w, response)
}

const maxFormSizeBytes = 32 << 20 // 32 megabytes

// UploadFile handles file uploads
// @Summary Upload a file
// @Description Uploads a new file. Admin users create common files, regular users create files owned by themselves.
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param file formData file true "File to upload"
// @Success 201 {object} FileResponse
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error
// @Router /files [post]
// @Security BearerAuth
func (a *API) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/upload_file")

	// Get identity from context
	identity, ok := GetIdentityFromContext(ctx)
	if !ok {
		a.writeJSON(ctx, w, respond.WithError(ctx, iam.ErrUnauthorized))
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(maxFormSizeBytes)
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}
	defer file.Close()

	opts := sesc.FileCreateOptions{
		FileName: header.Filename,
		FileSize: int(header.Size),
	}

	// Set owner ID based on user role:
	// - Admin users create common files (no owner)
	// - Regular users create files with themselves as owner
	isAdmin := identity.Role == iam.Role("admin")
	if !isAdmin {
		// Regular user - set owner to current user
		opts.OwnerID = &identity.ID
	}

	// Create the file
	newFile, err := a.file.Create(ctx, file, opts)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := convertFile(newFile)

	a.writeJSON(ctx, w, respond.WithStatus(response, http.StatusCreated))
}

// GetFileByID returns a file by ID
// @Summary Get file by ID
// @Description Returns a file by ID with download URL
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "File ID"
// @Success 200 {object} FileResponse
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /files/{id} [get]
// @Security BearerAuth
func (a *API) GetFileByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/get_file_by_id")

	fileID, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	file, err := a.file.ByID(ctx, fileID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	response := convertFile(file)

	a.writeJSON(ctx, w, response)
}

// DeleteFile deletes a file
// @Summary Delete file
// @Description Deletes a file by ID
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "File ID"
// @Success 204 "No Content"
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /files/{id} [delete]
// @Security BearerAuth
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/delete_file")

	fileID, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// FileAccessMiddleware has already checked access permissions
	err = a.file.Delete(ctx, fileID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
