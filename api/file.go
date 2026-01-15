package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

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

	user := CurrentUser(ctx)

	if ownerIDStr := query.Get("owner_id"); ownerIDStr != "" {
		if ownerIDStr == "me" {
			ownerIDStr = user.ID
		}

		opts.OwnerID = &ownerIDStr
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

	files, totalCount, err := a.file.Search(ctx, user, opts)
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
// @Param common query bool false "If true, upload as a common file"
// @Success 201 {object} respond.File
// @Failure 400 {object} respond.Error
// @Failure 401 {object} respond.Error "Unauthorized"
// @Failure 500 {object} respond.Error
// @Router /files [post]
// @Security BearerAuth
func (a *API) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/upload_file")

	isCommon := false
	if v := r.URL.Query().Get("common"); v != "" {
		common, err := strconv.ParseBool(v)
		if err != nil {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		isCommon = common
	}
	rec.Set("common_param", isCommon)

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

	user := CurrentUser(ctx)

	opts := sesc.FileCreateOptions{
		FileName: header.Filename,
		FileSize: int(header.Size),
		Common:   isCommon,
	}

	// Create the file
	newFile, err := a.file.Create(ctx, user, file, opts)
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
	user := CurrentUser(ctx)

	file, err := a.file.ByID(ctx, user, fileID)
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

	user := CurrentUser(ctx)
	err = a.file.Delete(ctx, user, fileID)
	if err != nil {
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadFile returns a pre-signed URL for file download
// @Summary Download file
// @Description Returns a pre-signed URL for downloading the file
// @Tags files
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer JWT token"
// @Param id path string true "File ID"
// @Param redirect query bool false "If true, redirect to the URL instead of returning JSON"
// @Success 200 {object} map[string]string "Returns {url: string} with pre-signed download URL"
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

	user := CurrentUser(ctx)
	// Get pre-signed download URL
	downloadURL, err := a.file.DownloadURL(ctx, user, fileID)
	if err != nil {
		if errors.Is(err, sesc.ErrFileNotFound) {
			a.writeJSON(ctx, w, respond.WithError(ctx, err))
			return
		}
		rec.Add(events.Error, err)
		a.writeJSON(ctx, w, respond.WithError(ctx, err))
		return
	}

	// Check if redirect is requested (for direct links like <a href> or <img src>)
	if r.URL.Query().Get("redirect") == "true" {
		http.Redirect(w, r, downloadURL, http.StatusTemporaryRedirect)
		return
	}

	a.writeJSON(ctx, w, map[string]string{"url": downloadURL})
}
