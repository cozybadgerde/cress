package serve

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A wrapped ResponseWriter hides the optional interfaces the real one
// implements. http.ResponseController finds them again through Unwrap, and its
// absence would surface as a flush that silently does nothing rather than as a
// build failure, so the contract is worth pinning.
func TestNotFoundWriterUnwraps(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &notFoundWriter{ResponseWriter: rec, outPath: "unused"}

	if err := http.NewResponseController(w).Flush(); err != nil {
		t.Fatalf("flushing through the wrapper: %v", err)
	}
	if !rec.Flushed {
		t.Error("Flush did not reach the underlying writer")
	}
}
