package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/storage"
	"github.com/google/uuid"
)

// UploadHandler manages image uploads for the admin dashboard.
type UploadHandler struct {
	Storage storage.StorageDriver
}

// Upload receives a multipart file and saves it using the configured cloud driver.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	logger.InfoCtx(r.Context(), "Entering UploadHandler.Upload - Content-Length: %d", r.ContentLength)
	if h.Storage == nil {
		logger.ErrorCtx(r.Context(), "Upload failed: Storage driver is not initialized. Check STORAGE_TYPE and credentials.")
		http.Error(w, "Storage not configured", http.StatusInternalServerError)
		return
	}

	// Limit upload size to 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Upload too large or invalid (Max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Simple file validation
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !validExts[ext] {
		http.Error(w, "Unsupported file type (jpg, png, webp only)", http.StatusBadRequest)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // Fallback
	}

	// Generate a unique filename using UUID to avoid collisions
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	ctx := r.Context()
	url, err := h.Storage.Upload(ctx, newFileName, contentType, file)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to upload file %s: %v", newFileName, err)
		http.Error(w, fmt.Sprintf("Failed to upload: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the public URL as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"url": %q}`, url)))
}

// NewUploadHandler creates a new handler instance.
func NewUploadHandler(s storage.StorageDriver) *UploadHandler {
	return &UploadHandler{Storage: s}
}
