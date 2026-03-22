package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/infrastructure/s3"
	"github.com/google/uuid"
)

// S3Presigner abstracts the S3 client for generating upload URLs.
type S3Presigner interface {
	GeneratePresignedURL(ctx context.Context, fileKey, mimeType string) (*s3.PresignedURL, error)
}

// UploadHandler handles media upload flow.
type UploadHandler struct {
	s3 S3Presigner
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(s3Client S3Presigner) *UploadHandler {
	return &UploadHandler{s3: s3Client}
}

type uploadURLRequest struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type uploadURLResponse struct {
	UploadURL string    `json:"upload_url"`
	FileKey   string    `json:"file_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RequestUploadURL handles POST /api/upload/request.
func (h *UploadHandler) RequestUploadURL(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req uploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s3.ValidateUpload(req.MimeType, req.SizeBytes); err != nil {
		switch {
		case errors.Is(err, s3.ErrInvalidMimeType):
			writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported file type: %s", req.MimeType))
		case errors.Is(err, s3.ErrFileTooLarge):
			writeError(w, http.StatusBadRequest, "file size exceeds 10MB limit")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	ext := extensionForMime(req.MimeType)
	fileKey := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)

	presigned, err := h.s3.GeneratePresignedURL(r.Context(), fileKey, req.MimeType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	writeJSON(w, http.StatusOK, uploadURLResponse{
		UploadURL: presigned.UploadURL,
		FileKey:   presigned.FileKey,
		ExpiresAt: presigned.ExpiresAt,
	})
}

func extensionForMime(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
