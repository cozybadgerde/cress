package preview

import (
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

const rootRoute = "/"

type pageServer struct {
	mu     sync.RWMutex
	root   *os.Root
	prefix string
}

type publishedFS struct{ source fs.FS }

// replaceRoot requires the caller to hold mu through publication and replacement.
func (server *pageServer) replaceRoot(directory string) error {
	next, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	old := server.root
	server.root = next
	return old.Close()
}

func (files publishedFS) Open(name string) (fs.File, error) {
	if strings.EqualFold(name, ".cress-output.json") {
		return nil, fs.ErrNotExist
	}
	file, err := files.source.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if info.IsDir() {
		index, err := files.source.Open(path.Join(name, "index.html"))
		if err != nil {
			return nil, errors.Join(err, file.Close())
		}
		info, statErr := index.Stat()
		if err := errors.Join(statErr, index.Close()); err != nil {
			return nil, errors.Join(err, file.Close())
		}
		if !info.Mode().IsRegular() {
			return nil, errors.Join(fs.ErrNotExist, file.Close())
		}
	}
	return file, nil
}

func (server *pageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	if server.root == nil {
		http.Error(w, "preview output is unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if server.prefix != rootRoute && r.URL.Path == rootRoute {
		http.Redirect(w, r, server.prefix, http.StatusTemporaryRedirect)
		return
	}
	decodedPrefix, err := url.PathUnescape(server.prefix)
	if err != nil {
		http.Error(w, "invalid preview prefix", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(r.URL.Path, decodedPrefix) {
		http.NotFound(w, r)
		return
	}
	request := r.Clone(r.Context())
	copiedURL := *r.URL
	request.URL = &copiedURL
	request.URL.Path = rootRoute + strings.TrimPrefix(r.URL.Path, decodedPrefix)
	request.URL.RawPath = ""
	http.FileServerFS(publishedFS{source: server.root.FS()}).ServeHTTP(w, request)
}
