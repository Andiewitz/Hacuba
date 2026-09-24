package images

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type pendingUpload struct {
	contentType string
	byteSize    int64
}

// LocalStore is a development-only object store. It implements the same
// presign/register contract as S3, but writes to a local directory through a
// short-lived expected-upload record. Production never constructs this type.
type LocalStore struct {
	root      string
	publicURL string
	mu        sync.Mutex
	pending   map[string]pendingUpload
}

func NewLocalStore(root, publicURL string) (*LocalStore, error) {
	if root == "" || publicURL == "" {
		return nil, fmt.Errorf("local upload root and public URL are required")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create local upload directory: %w", err)
	}
	return &LocalStore{root: root, publicURL: strings.TrimRight(publicURL, "/"), pending: map[string]pendingUpload{}}, nil
}

func (s *LocalStore) PresignPut(_ context.Context, key, contentType string, size int64) (string, error) {
	s.mu.Lock()
	s.pending[key] = pendingUpload{contentType: contentType, byteSize: size}
	s.mu.Unlock()
	return s.publicURL + "/" + key, nil
}

func (s *LocalStore) Head(_ context.Context, key string) error {
	_, err := os.Stat(s.pathFor(key))
	return err
}

func (s *LocalStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/dev-uploads/")
	path, ok := s.safePath(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPut:
		s.put(w, r, key, path)
	case http.MethodGet:
		http.ServeFile(w, r, path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *LocalStore) put(w http.ResponseWriter, r *http.Request, key, path string) {
	s.mu.Lock()
	expected, exists := s.pending[key]
	s.mu.Unlock()
	if !exists || r.Header.Get("Content-Type") != expected.contentType || (r.ContentLength >= 0 && r.ContentLength != expected.byteSize) {
		http.Error(w, "invalid development upload", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		http.Error(w, "upload unavailable", http.StatusInternalServerError)
		return
	}
	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "upload unavailable", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	written, err := io.Copy(file, http.MaxBytesReader(w, r.Body, expected.byteSize+1))
	if err != nil || written != expected.byteSize {
		_ = os.Remove(path)
		http.Error(w, "invalid development upload", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	delete(s.pending, key)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *LocalStore) pathFor(key string) string {
	path, _ := s.safePath(key)
	return path
}

func (s *LocalStore) safePath(key string) (string, bool) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	path := filepath.Join(s.root, clean)
	relative, err := filepath.Rel(s.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return path, true
}
