package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
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

// FileNotFoundError represents a file not found error
type FileNotFoundError struct {
	Code       string `json:"code"             example:"FILE_NOT_FOUND"`
	Message    string `json:"message"          example:"File not found"`
	RuMessage  string `json:"ruMessage"        example:"Файл не найден"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e FileNotFoundError) WithDetails(details string) FileNotFoundError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e FileNotFoundError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// FileForbiddenError represents an access denied error for file operations
type FileForbiddenError struct {
	Code       string `json:"code"             example:"FILE_ACCESS_DENIED"`
	Message    string `json:"message"          example:"Access to file denied"`
	RuMessage  string `json:"ruMessage"        example:"Доступ к файлу запрещен"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e FileForbiddenError) WithDetails(details string) FileForbiddenError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e FileForbiddenError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// FileOperationError represents a general error during file operations
type FileOperationError struct {
	Code       string `json:"code"             example:"FILE_OPERATION_ERROR"`
	Message    string `json:"message"          example:"File operation failed"`
	RuMessage  string `json:"ruMessage"        example:"Операция с файлом не удалась"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e FileOperationError) WithDetails(details string) FileOperationError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e FileOperationError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

// BadRequestError represents a bad request error
type BadRequestError struct {
	Code       string `json:"code"             example:"BAD_REQUEST"`
	Message    string `json:"message"          example:"Bad request"`
	RuMessage  string `json:"ruMessage"        example:"Неверный запрос"`
	Details    string `json:"details,omitzero"`
	StatusCode int    `json:"-"`
}

// WithDetails adds detail information to the error
func (e BadRequestError) WithDetails(details string) BadRequestError {
	e.Details = details
	return e
}

// WithStatus adds HTTP status code to the error
func (e BadRequestError) WithStatus(statusCode int) Error {
	e.StatusCode = statusCode
	return Error(e)
}

var (
	ErrBadRequest = BadRequestError{
		Code:      "BAD_REQUEST",
		Message:   "Bad request",
		RuMessage: "Неверный запрос",
	}

	ErrNotFound = FileNotFoundError{
		Code:      "FILE_NOT_FOUND",
		Message:   "File not found",
		RuMessage: "Файл не найден",
	}

	ErrFileForbidden = FileForbiddenError{
		Code:      "FILE_ACCESS_DENIED",
		Message:   "Access to file denied",
		RuMessage: "Доступ к файлу запрещен",
	}

	ErrFileOperation = FileOperationError{
		Code:      "FILE_OPERATION_ERROR",
		Message:   "File operation failed",
		RuMessage: "Операция с файлом не удалась",
	}

	ErrInvalidFileUpload = FileOperationError{
		Code:      "INVALID_FILE_UPLOAD",
		Message:   "Invalid file upload",
		RuMessage: "Некорректная загрузка файла",
	}

	ErrUnauthorizedFileDelete = FileForbiddenError{
		Code:      "UNAUTHORIZED_FILE_DELETE",
		Message:   "You are not authorized to delete this file",
		RuMessage: "Вы не можете удалить этот файл",
	}
)

// internalError converts any error to a ServerError
func internalError(err error) Error {
	return ServerError{
		Code:      "SERVER_ERROR",
		Message:   "Internal server error",
		RuMessage: "Внутренняя ошибка сервера",
		Details:   err.Error(),
	}.WithStatus(http.StatusInternalServerError)
}

// fileError converts file-specific domain errors to API errors
func fileError(err error) Error {
	switch {
	case errors.Is(err, sesc.ErrFileNotFound):
		return ErrNotFound.WithStatus(http.StatusNotFound)
	case errors.Is(err, sesc.ErrInvalidFileName):
		return ErrBadRequest.WithDetails("Invalid file name").WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrInvalidFileSize):
		return ErrBadRequest.WithDetails("Invalid file size").WithStatus(http.StatusBadRequest)
	case errors.Is(err, sesc.ErrInvalidFile):
		return ErrBadRequest.WithDetails("Invalid file").WithStatus(http.StatusBadRequest)
	default:
		return internalError(err)
	}
}

// FileAccessMiddleware checks if the user has access to the requested file
func (a *API) FileAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx).Sub("file_access_check")

		// Extract file ID from URL
		fileID, err := uuid.FromString(chi.URLParam(r, "id"))
		if err != nil {
			writeError(ctx, w, ErrBadRequest.WithDetails("invalid file ID").WithStatus(http.StatusBadRequest))
			return
		}

		// Get identity from context
		identity, identityOk := GetIdentityFromContext(ctx)
		if !identityOk {
			writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
			return
		}

		// Get file details to check ownership
		file, err := a.file.ByID(ctx, fileID)
		if err != nil {
			if errors.Is(err, sesc.ErrFileNotFound) {
				writeError(ctx, w, ErrNotFound.WithDetails("file not found").WithStatus(http.StatusNotFound))
				return
			}
			rec.Add(events.Error, err)
			writeError(ctx, w, fileError(err))
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
		writeError(ctx, w, ErrFileForbidden.WithStatus(http.StatusForbidden))
	})
}

func (a *API) FileEditAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx).Sub("file_access_check")

		// Extract file ID from URL
		fileID, err := uuid.FromString(chi.URLParam(r, "id"))
		if err != nil {
			writeError(ctx, w, ErrBadRequest.WithDetails("invalid file ID").WithStatus(http.StatusBadRequest))
			return
		}

		// Get identity from context
		identity, identityOk := GetIdentityFromContext(ctx)
		if !identityOk {
			writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
			return
		}

		// Get file details to check ownership
		file, err := a.file.ByID(ctx, fileID)
		if err != nil {
			if errors.Is(err, sesc.ErrFileNotFound) {
				writeError(ctx, w, ErrNotFound.WithDetails("file not found").WithStatus(http.StatusNotFound))
				return
			}
			rec.Add(events.Error, err)
			writeError(ctx, w, fileError(err))
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
		writeError(ctx, w, ErrFileForbidden.WithStatus(http.StatusForbidden))
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
// @Failure 400 {object} Error
// @Failure 500 {object} Error
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
				writeError(ctx, w, ErrBadRequest.WithDetails("invalid owner_id").WithStatus(http.StatusBadRequest))
				return
			}
			opts.OwnerID = &ownerID
		}
	}

	if commonStr := query.Get("common"); commonStr != "" {
		common, err := strconv.ParseBool(commonStr)
		if err != nil {
			writeError(ctx, w, ErrBadRequest.WithDetails("invalid common parameter").WithStatus(http.StatusBadRequest))
			return
		}
		opts.Common = common
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			writeError(ctx, w, ErrBadRequest.WithDetails("invalid offset").WithStatus(http.StatusBadRequest))
			return
		}
		opts.Offset = offset
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			writeError(ctx, w, ErrBadRequest.WithDetails("invalid limit").WithStatus(http.StatusBadRequest))
			return
		}
		opts.Limit = limit
	}

	files, totalCount, err := a.file.Search(ctx, opts)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, fileError(err))
		return
	}

	response := FileListResponse{
		Items:      make([]FileResponse, len(files)),
		TotalCount: totalCount,
	}

	for i, file := range files {
		response.Items[i] = convertFile(file)
	}

	a.writeJSON(ctx, w, response, http.StatusOK)
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
// @Failure 400 {object} Error
// @Failure 401 {object} Error "Unauthorized"
// @Failure 500 {object} Error
// @Router /files [post]
// @Security BearerAuth
func (a *API) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/upload_file")

	// Get identity from context
	identity, ok := GetIdentityFromContext(ctx)
	if !ok {
		writeError(ctx, w, ErrUnauthorized.WithStatus(http.StatusUnauthorized))
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(maxFormSizeBytes)
	if err != nil {
		writeError(ctx, w, ErrInvalidFileUpload.WithDetails("unable to parse form").WithStatus(http.StatusBadRequest))
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(ctx, w, ErrInvalidFileUpload.WithDetails("file field is required").WithStatus(http.StatusBadRequest))
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
		writeError(ctx, w, fileError(err))
		return
	}

	response := convertFile(newFile)

	a.writeJSON(ctx, w, response, http.StatusCreated)
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
// @Failure 400 {object} Error
// @Failure 401 {object} Error "Unauthorized"
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /files/{id} [get]
// @Security BearerAuth
func (a *API) GetFileByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/get_file_by_id")

	fileID, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		writeError(ctx, w, ErrBadRequest.WithDetails("invalid file ID").WithStatus(http.StatusBadRequest))
		return
	}

	file, err := a.file.ByID(ctx, fileID)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, fileError(err))
		return
	}

	response := convertFile(file)

	a.writeJSON(ctx, w, response, http.StatusOK)
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
// @Failure 400 {object} Error
// @Failure 401 {object} Error "Unauthorized"
// @Failure 403 {object} Error "Forbidden"
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /files/{id} [delete]
// @Security BearerAuth
func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rec := event.Get(ctx).Sub("api/delete_file")

	fileID, err := uuid.FromString(chi.URLParam(r, "id"))
	if err != nil {
		writeError(ctx, w, ErrBadRequest.WithDetails("invalid file ID").WithStatus(http.StatusBadRequest))
		return
	}

	// FileAccessMiddleware has already checked access permissions
	err = a.file.Delete(ctx, fileID)
	if err != nil {
		rec.Add(events.Error, err)
		writeError(ctx, w, fileError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
