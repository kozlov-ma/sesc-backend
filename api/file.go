package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// FileAccessMiddleware checks if the user has access to the requested file
func (a *API) FileAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx).Sub("file_access_check")

		// Extract file ID from URL

		fileID, err := param.PathUUID(r, "id")

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

		fileID, err := param.PathUUID(r, "id")

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
// @Success 200 {object} respond.Files
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
		rec.Add("common_param", common)
	} else {
		rec.Add("common_param", "not_provided")
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

	a.writeJSON(ctx, w, respond.WithFiles(files, totalCount))
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
// @Success 201 {object} respond.File
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

	a.writeJSON(ctx, w, respond.WithStatus(respond.WithFile(newFile), http.StatusCreated))
}

// GetFileByID returns a file by ID
// @Summary Get file by ID
// @Description Returns a file by ID with download URL
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "File ID"
// @Success 200 {object} respond.File
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 404 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /files/{id} [get]
// @Security BearerAuth
func (a *API) GetFileByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/get_file_by_id")

	fileID, err := param.PathUUID(r, "id")

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

	a.writeJSON(ctx, w, respond.WithFile(file))
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

	fileID, err := param.PathUUID(r, "id")

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

// DownloadFile redirects to a pre-signed URL for file download
// @Summary Download file
// @Description Redirects to a pre-signed URL for downloading the file
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "File ID"
// @Success 307
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 403 {object} respond.Error "Forbidden"
// @Failure 404 {object} respond.Error
// @Failure 500 {object} respond.Error
// @Router /files/{id}/download [get]
// @Security BearerAuth
func (a *API) DownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/download_file")

	// Extract file ID from URL
	fileID, err := param.PathUUID(r, "id")
	if err != nil {
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Get pre-signed download URL
	downloadURL, err := a.file.DownloadURL(ctx, fileID)
	if err != nil {
		if errors.Is(err, sesc.ErrFileNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Redirect to pre-signed URL
	http.Redirect(w, r, downloadURL, http.StatusTemporaryRedirect)
}
