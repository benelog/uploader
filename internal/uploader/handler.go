// Package uploader serves a small web form that stores an uploaded file
// into a directory chosen by the user.
package uploader

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed templates/*.html
var templateFS embed.FS

const (
	formViewName = "uploadForm.html"

	// maxFieldSize bounds the plain text fields of the form, so a malformed
	// request cannot force the server to buffer an unbounded value.
	maxFieldSize = 1 << 20
)

// ErrPathAfterFile is returned when the file part arrives before the server
// path, which leaves the destination directory unknown while the body is
// being streamed.
var ErrPathAfterFile = errors.New("'path' must be sent before 'file'")

// Handler renders the upload form and writes uploaded files to disk.
type Handler struct {
	templates *template.Template
}

// NewHandler parses the embedded templates.
func NewHandler() (*Handler, error) {
	templates, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Handler{templates: templates}, nil
}

// Routes maps the URLs kept from the original servlet mapping (*.html).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/form.html", h.Form)
	mux.HandleFunc("/upload.html", h.Upload)
	mux.HandleFunc("/", redirectToForm)
	return mux
}

func redirectToForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/form.html", http.StatusFound)
}

// Form shows an empty upload form.
func (h *Handler) Form(w http.ResponseWriter, _ *http.Request) {
	h.render(w, "ready!")
}

// Upload streams the posted file to the requested server path. The request
// body is never buffered as a whole, so the size of an upload is limited by
// the destination disk rather than by memory or by a temporary copy.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	message, err := h.storeRequest(r)
	switch {
	case errors.Is(err, ErrPathAfterFile):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, errNoFilePart):
		http.Error(w, "required parameter 'file' is not present", http.StatusBadRequest)
		return
	case errors.Is(err, errBadRequest):
		http.Error(w, "invalid multipart request", http.StatusBadRequest)
		return
	case err != nil:
		log.Printf("upload failed: %v", err)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	log.Print(message)
	h.render(w, message)
}

var (
	errNoFilePart = errors.New("no file part")
	errBadRequest = errors.New("malformed multipart request")
)

// storeRequest walks the multipart body part by part and writes the file part
// straight to its destination.
func (h *Handler) storeRequest(r *http.Request) (string, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", fmt.Errorf("%w: %v", errBadRequest, err)
	}

	var (
		path    string
		sawPath bool
		message string
		stored  bool
	)
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("%w: %v", errBadRequest, err)
		}

		switch part.FormName() {
		case "path":
			value, err := io.ReadAll(io.LimitReader(part, maxFieldSize))
			_ = part.Close()
			if err != nil {
				return "", fmt.Errorf("%w: %v", errBadRequest, err)
			}
			path, sawPath = string(value), true
		case "file":
			if !sawPath {
				_ = part.Close()
				return "", ErrPathAfterFile
			}
			message, err = Store(part, part.FileName(), path)
			_ = part.Close()
			if err != nil {
				return "", err
			}
			stored = true
		default:
			_ = part.Close()
		}
	}

	if !stored {
		return "", errNoFilePart
	}
	return message, nil
}

// Store copies file into path under its own name and returns the message
// shown to the user. The content is streamed, so a file of any size only
// needs room at the destination. An empty file is reported instead of being
// written, as in the original implementation.
func Store(file io.Reader, filename, path string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("%w: the file part has no filename", errBadRequest)
	}

	// Detecting emptiness before creating the file keeps an existing file at
	// the destination untouched when an empty upload is submitted.
	buffered := bufio.NewReader(file)
	if _, err := buffered.Peek(1); errors.Is(err, io.EOF) {
		return "empty file", nil
	} else if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}

	// The submitted name is only a hint: keep the last element so a crafted
	// name cannot escape the requested directory.
	dest := filepath.Join(path, filepath.Base(filename))
	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", dest, err)
	}
	// Safety net for the error paths below; the successful path closes the
	// file explicitly to report a failed flush.
	defer func() { _ = out.Close() }()

	written, err := io.Copy(out, buffered)
	if err != nil {
		_ = os.Remove(dest)
		return "", fmt.Errorf("write %s: %w", dest, err)
	}
	if err := out.Close(); err != nil {
		return "", fmt.Errorf("close %s: %w", dest, err)
	}

	return createMessage(dest, written), nil
}

func createMessage(dest string, size int64) string {
	writtenPath, err := filepath.Abs(dest)
	if err != nil {
		writtenPath = dest
	}
	return fmt.Sprintf("%s (%dbytes) uploaded!", writtenPath, size)
}

func (h *Handler) render(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if err := h.templates.ExecuteTemplate(w, formViewName, map[string]any{"message": message}); err != nil {
		log.Printf("render failed: %v", err)
	}
}
