package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAllowedMimeType_ImagePNG_ReturnsTrue(t *testing.T) {
	assert.True(t, IsAllowedMimeType("image/png"))
}

func TestIsAllowedMimeType_ImageJPEG_ReturnsTrue(t *testing.T) {
	assert.True(t, IsAllowedMimeType("image/jpeg"))
}

func TestIsAllowedMimeType_Executable_ReturnsFalse(t *testing.T) {
	assert.False(t, IsAllowedMimeType("application/x-executable"))
}

func TestIsAllowedMimeType_PDF_ReturnsFalse(t *testing.T) {
	assert.False(t, IsAllowedMimeType("application/pdf"))
}

func TestValidateUpload_ValidImageSmall_ReturnsNil(t *testing.T) {
	err := ValidateUpload("image/png", 1024)
	assert.NoError(t, err)
}

func TestValidateUpload_InvalidMimeType_ReturnsError(t *testing.T) {
	err := ValidateUpload("application/zip", 1024)
	assert.ErrorIs(t, err, ErrInvalidMimeType)
}

func TestValidateUpload_FileTooLarge_ReturnsError(t *testing.T) {
	err := ValidateUpload("image/png", MaxFileSizeBytes+1)
	assert.ErrorIs(t, err, ErrFileTooLarge)
}

func TestValidateUpload_ExactlyMaxSize_ReturnsNil(t *testing.T) {
	err := ValidateUpload("image/jpeg", MaxFileSizeBytes)
	assert.NoError(t, err)
}
