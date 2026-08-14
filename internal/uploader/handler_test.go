package uploader

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileShouldBeUploaded(t *testing.T) {
	// given
	fileName := "test.txt"
	fileContent := "my works"
	serverPath := t.TempDir()

	// when
	message, err := Store(strings.NewReader(fileContent), fileName, serverPath)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// then
	uploaded, err := os.ReadFile(filepath.Join(serverPath, fileName))
	if err != nil {
		t.Fatalf("uploaded file is not readable: %v", err)
	}
	if string(uploaded) != fileContent {
		t.Errorf("uploaded content = %q, want %q", uploaded, fileContent)
	}
	if message == "" {
		t.Error("message should not be empty")
	}
	t.Log(message)
}

func TestEmptyFileIsNotWritten(t *testing.T) {
	serverPath := t.TempDir()
	existing := filepath.Join(serverPath, "empty.txt")
	if err := os.WriteFile(existing, []byte("previous content"), 0o600); err != nil {
		t.Fatalf("prepare existing file: %v", err)
	}

	message, err := Store(strings.NewReader(""), "empty.txt", serverPath)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if message != "empty file" {
		t.Errorf("message = %q, want %q", message, "empty file")
	}
	kept, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("existing file is not readable: %v", err)
	}
	if string(kept) != "previous content" {
		t.Errorf("an empty upload must not truncate an existing file, got %q", kept)
	}
}

func TestFileNameCannotEscapeTheServerPath(t *testing.T) {
	serverPath := t.TempDir()

	if _, err := Store(strings.NewReader("x"), "../escaped.txt", serverPath); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(serverPath, "escaped.txt")); err != nil {
		t.Errorf("file should be written inside the server path: %v", err)
	}
}

func TestStoreRejectsAnEmptyFileName(t *testing.T) {
	if _, err := Store(strings.NewReader("x"), "", t.TempDir()); !errors.Is(err, errBadRequest) {
		t.Errorf("err = %v, want errBadRequest", err)
	}
}

func TestUploadHandlerRendersTheForm(t *testing.T) {
	serverPath := t.TempDir()
	handler := newTestHandler(t)

	body, contentType := multipartBody(t, serverPath, "file", "test.txt", "my works")
	rec := serve(handler, newUploadRequest(body, contentType))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !strings.Contains(got, "uploaded!") {
		t.Errorf("response should report the upload, got:\n%s", got)
	}
	uploaded, err := os.ReadFile(filepath.Join(serverPath, "test.txt"))
	if err != nil {
		t.Fatalf("uploaded file is not readable: %v", err)
	}
	if string(uploaded) != "my works" {
		t.Errorf("uploaded content = %q, want %q", uploaded, "my works")
	}
}

// TestLargeFileIsStreamed uploads more than any in-memory buffer would hold,
// to prove the body is written straight to the destination.
func TestLargeFileIsStreamed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large upload in short mode")
	}

	const size = 256 << 20 // 256 MiB
	serverPath := t.TempDir()
	handler := newTestHandler(t)

	// The body is produced on the fly so the test itself never holds the
	// payload in memory either.
	body, contentType := streamingMultipartBody(t, serverPath, "big.bin", size)
	rec := serve(handler, newUploadRequest(body, contentType))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body)
	}
	info, err := os.Stat(filepath.Join(serverPath, "big.bin"))
	if err != nil {
		t.Fatalf("uploaded file is not readable: %v", err)
	}
	if info.Size() != size {
		t.Errorf("uploaded size = %d, want %d", info.Size(), size)
	}
	if !strings.Contains(rec.Body.String(), "(268435456bytes) uploaded!") {
		t.Errorf("message should report the full size, got:\n%s", rec.Body)
	}
}

func TestUploadWithoutFileIsRejected(t *testing.T) {
	handler := newTestHandler(t)

	body, contentType := multipartBody(t, t.TempDir(), "", "", "")
	rec := serve(handler, newUploadRequest(body, contentType))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFileBeforePathIsRejected(t *testing.T) {
	handler := newTestHandler(t)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.WriteString(part, "my works"); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.WriteField("path", t.TempDir()); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	rec := serve(handler, newUploadRequest(&buf, writer.FormDataContentType()))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "must be sent before") {
		t.Errorf("response should explain the part order, got: %s", rec.Body)
	}
}

func TestGetUploadIsNotAllowed(t *testing.T) {
	handler := newTestHandler(t)

	rec := serve(handler, httptest.NewRequest(http.MethodGet, "/upload.html", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestFormShowsReadyMessage(t *testing.T) {
	handler := newTestHandler(t)

	rec := serve(handler, httptest.NewRequest(http.MethodGet, "/form.html", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "ready!") {
		t.Error("form should show the ready message")
	}
}

func TestRootRedirectsToForm(t *testing.T) {
	handler := newTestHandler(t)

	rec := serve(handler, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); location != "/form.html" {
		t.Errorf("Location = %q, want %q", location, "/form.html")
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	handler := newTestHandler(t)

	rec := serve(handler, httptest.NewRequest(http.MethodGet, "/nowhere", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}
	return handler
}

func serve(handler *Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)
	return rec
}

func newUploadRequest(body io.Reader, contentType string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/upload.html", body)
	req.Header.Set("Content-Type", contentType)
	return req
}

// multipartBody builds a multipart body with the path field first, as the
// HTML form does. The file part is omitted when fieldName is empty.
func multipartBody(t *testing.T, path, fieldName, fileName, content string) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("path", path); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if fieldName != "" {
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := io.WriteString(part, content); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &buf, writer.FormDataContentType()
}

// streamingMultipartBody returns a body that generates size bytes of file
// content as it is read.
func streamingMultipartBody(t *testing.T, path, fileName string, size int64) (io.Reader, string) {
	t.Helper()
	reader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	go func() {
		err := func() error {
			if err := writer.WriteField("path", path); err != nil {
				return err
			}
			part, err := writer.CreateFormFile("file", fileName)
			if err != nil {
				return err
			}
			if _, err := io.CopyN(part, newFillReader('a'), size); err != nil {
				return err
			}
			return writer.Close()
		}()
		_ = pipeWriter.CloseWithError(err)
	}()

	return reader, writer.FormDataContentType()
}

// fillReader is an endless source of one repeated byte.
type fillReader struct{ b byte }

func newFillReader(b byte) io.Reader { return fillReader{b: b} }

func (f fillReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = f.b
	}
	return len(p), nil
}
