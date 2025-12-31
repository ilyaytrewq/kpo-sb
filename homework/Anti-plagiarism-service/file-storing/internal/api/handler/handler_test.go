package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/service"
)

type uploadCall struct {
	bucket           string
	key              string
	contentType      string
	originalFileName string
	size             int64
}

type fakeObjectService struct {
	bucket string

	uploadErr error
	uploadLog []uploadCall

	downloadErr         error
	downloadContent     []byte
	downloadContentType string

	headErr  error
	headInfo service.Info
}

func (f *fakeObjectService) Bucket() string {
	return f.bucket
}

func (f *fakeObjectService) Upload(ctx context.Context, bucket, key, contentType, originalFileName string, body io.Reader, size int64) error {
	if body != nil {
		_, _ = io.ReadAll(body)
	}
	f.uploadLog = append(f.uploadLog, uploadCall{
		bucket:           bucket,
		key:              key,
		contentType:      contentType,
		originalFileName: originalFileName,
		size:             size,
	})
	return f.uploadErr
}

func (f *fakeObjectService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, string, error) {
	if f.downloadErr != nil {
		return nil, "", f.downloadErr
	}
	return io.NopCloser(bytes.NewReader(f.downloadContent)), f.downloadContentType, nil
}

func (f *fakeObjectService) Head(ctx context.Context, bucket, key string) (service.Info, error) {
	if f.headErr != nil {
		return service.Info{}, f.headErr
	}
	return f.headInfo, nil
}

func TestUploadFile_Success(t *testing.T) {
	fakeSvc := &fakeObjectService{
		bucket: "bucket-1",
	}
	h := &Handler{service: fakeSvc}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("metadata", `{"workId":"work-1","originalFileName":"hw3.pdf","contentType":"application/pdf"}`); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	part, err := writer.CreateFormFile("file", "original.pdf")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.UploadFile(rec, req, api.UploadFileParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp api.UploadFileResult
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.FileId == uuid.Nil {
		t.Fatalf("expected non-empty fileId")
	}
	if resp.ContentType == nil || *resp.ContentType != "application/pdf" {
		t.Fatalf("unexpected contentType: %#v", resp.ContentType)
	}
	if resp.OriginalFileName == nil || *resp.OriginalFileName != "hw3.pdf" {
		t.Fatalf("unexpected originalFileName: %#v", resp.OriginalFileName)
	}
	if resp.SizeBytes == nil || *resp.SizeBytes != int64(5) {
		t.Fatalf("unexpected sizeBytes: %#v", resp.SizeBytes)
	}
	if len(fakeSvc.uploadLog) != 1 {
		t.Fatalf("expected 1 upload call, got %d", len(fakeSvc.uploadLog))
	}
	if fakeSvc.uploadLog[0].bucket != "bucket-1" {
		t.Fatalf("unexpected bucket: %s", fakeSvc.uploadLog[0].bucket)
	}
}

func TestUploadFile_InvalidMetadata(t *testing.T) {
	fakeSvc := &fakeObjectService{bucket: "bucket-1"}
	h := &Handler{service: fakeSvc}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("metadata", "{"); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	part, err := writer.CreateFormFile("file", "file.txt")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.UploadFile(rec, req, api.UploadFileParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUploadFile_MissingWorkID(t *testing.T) {
	fakeSvc := &fakeObjectService{bucket: "bucket-1"}
	h := &Handler{service: fakeSvc}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("metadata", `{"workId":""}`); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	part, err := writer.CreateFormFile("file", "file.txt")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.UploadFile(rec, req, api.UploadFileParams{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDownloadFile_Success(t *testing.T) {
	fakeSvc := &fakeObjectService{
		bucket:              "bucket-1",
		downloadContent:     []byte("data"),
		downloadContentType: "text/plain",
	}
	h := &Handler{service: fakeSvc}

	fileID := openapi_types.UUID(uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"))
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileID.String(), nil)
	rec := httptest.NewRecorder()

	h.DownloadFile(rec, req, fileID, api.DownloadFileParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("unexpected content-type: %s", rec.Header().Get("Content-Type"))
	}
	if rec.Body.String() != "data" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestGetFileInfo_Success(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	fakeSvc := &fakeObjectService{
		bucket: "bucket-1",
		headInfo: service.Info{
			Size:        123,
			ContentType: "application/pdf",
			StoredAt:    now,
			FileName:    "hw3.pdf",
		},
	}
	h := &Handler{service: fakeSvc}

	fileID := openapi_types.UUID(uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"))
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileID.String()+"/info", nil)
	rec := httptest.NewRecorder()

	h.GetFileInfo(rec, req, fileID, api.GetFileInfoParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp api.FileInfoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.FileId != uuid.UUID(fileID) {
		t.Fatalf("unexpected fileId: %s", resp.FileId.String())
	}
	if resp.SizeBytes == nil || *resp.SizeBytes != 123 {
		t.Fatalf("unexpected size: %#v", resp.SizeBytes)
	}
	if resp.ContentType == nil || *resp.ContentType != "application/pdf" {
		t.Fatalf("unexpected contentType: %#v", resp.ContentType)
	}
	if resp.OriginalFileName == nil || *resp.OriginalFileName != "hw3.pdf" {
		t.Fatalf("unexpected originalFileName: %#v", resp.OriginalFileName)
	}
}
