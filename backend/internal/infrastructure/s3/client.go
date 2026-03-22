package s3

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidMimeType = errors.New("unsupported file type")
	ErrFileTooLarge    = errors.New("file exceeds maximum size of 10MB")

	MaxFileSizeBytes int64 = 10 * 1024 * 1024 // 10 MB

	allowedMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
)

// PresignedURL holds a pre-signed upload URL and associated metadata.
type PresignedURL struct {
	UploadURL string
	FileKey   string
	ExpiresAt time.Time
}

// S3Client generates pre-signed upload URLs. The real implementation uses
// the AWS SDK; a mock is used in tests.
type S3Client interface {
	GeneratePresignedURL(ctx context.Context, fileKey, mimeType string) (*PresignedURL, error)
}

// IsAllowedMimeType reports whether mimeType is permitted for upload.
func IsAllowedMimeType(mimeType string) bool {
	return allowedMimeTypes[mimeType]
}

// ValidateUpload validates the mime type and file size.
func ValidateUpload(mimeType string, sizeBytes int64) error {
	if !IsAllowedMimeType(mimeType) {
		return fmt.Errorf("%w: %s", ErrInvalidMimeType, mimeType)
	}
	if sizeBytes > MaxFileSizeBytes {
		return ErrFileTooLarge
	}
	return nil
}
