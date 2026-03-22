package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/infrastructure/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockS3Presigner struct{ mock.Mock }

func (m *MockS3Presigner) GeneratePresignedURL(ctx context.Context, fileKey, mimeType string) (*s3.PresignedURL, error) {
	args := m.Called(ctx, fileKey, mimeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PresignedURL), args.Error(1)
}

func makePresignedURL() *s3.PresignedURL {
	return &s3.PresignedURL{
		UploadURL: "https://bucket.s3.amazonaws.com/upload-path",
		FileKey:   "uploads/abc.png",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
}

func TestUploadHandler_RequestURL_ValidPNG_Returns200(t *testing.T) {
	presigner := new(MockS3Presigner)
	h := NewUploadHandler(presigner)

	presigner.On("GeneratePresignedURL", mock.Anything, mock.AnythingOfType("string"), "image/png").
		Return(makePresignedURL(), nil)

	body := `{"filename":"photo.png","mime_type":"image/png","size_bytes":1024}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp uploadURLResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.UploadURL)
	assert.NotEmpty(t, resp.FileKey)
}

func TestUploadHandler_RequestURL_InvalidMimeType_Returns400(t *testing.T) {
	h := NewUploadHandler(new(MockS3Presigner))

	body := `{"filename":"virus.exe","mime_type":"application/x-executable","size_bytes":1024}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUploadHandler_RequestURL_FileTooLarge_Returns400(t *testing.T) {
	h := NewUploadHandler(new(MockS3Presigner))

	body := `{"filename":"huge.png","mime_type":"image/png","size_bytes":20971520}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUploadHandler_RequestURL_NotAuthenticated_Returns401(t *testing.T) {
	h := NewUploadHandler(new(MockS3Presigner))

	body := `{"filename":"photo.png","mime_type":"image/png","size_bytes":1024}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUploadHandler_RequestURL_InvalidJSON_Returns400(t *testing.T) {
	h := NewUploadHandler(new(MockS3Presigner))

	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader("bad json"))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUploadHandler_RequestURL_JPEG_Returns200(t *testing.T) {
	presigner := new(MockS3Presigner)
	h := NewUploadHandler(presigner)

	presigner.On("GeneratePresignedURL", mock.Anything, mock.AnythingOfType("string"), "image/jpeg").
		Return(makePresignedURL(), nil)

	body := `{"filename":"photo.jpg","mime_type":"image/jpeg","size_bytes":2048}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUploadHandler_RequestURL_GIF_Returns200(t *testing.T) {
	presigner := new(MockS3Presigner)
	h := NewUploadHandler(presigner)

	presigner.On("GeneratePresignedURL", mock.Anything, mock.AnythingOfType("string"), "image/gif").
		Return(makePresignedURL(), nil)

	body := `{"filename":"anim.gif","mime_type":"image/gif","size_bytes":4096}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUploadHandler_RequestURL_WEBP_Returns200(t *testing.T) {
	presigner := new(MockS3Presigner)
	h := NewUploadHandler(presigner)

	presigner.On("GeneratePresignedURL", mock.Anything, mock.AnythingOfType("string"), "image/webp").
		Return(makePresignedURL(), nil)

	body := `{"filename":"photo.webp","mime_type":"image/webp","size_bytes":2048}`
	req := httptest.NewRequest(http.MethodPost, "/api/upload/request", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RequestUploadURL(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
