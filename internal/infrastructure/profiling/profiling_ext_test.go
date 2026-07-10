package profiling_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/profiling"
)

// Black-box: WriteHeap produces a non-empty heap profile.
func TestWriteHeap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "heap.prof")
	if err := profiling.WriteHeap(path); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("heap profile not written: err=%v", err)
	}
}

// Black-box: WriteHeap surfaces file-creation errors.
func TestWriteHeapBadPath(t *testing.T) {
	if err := profiling.WriteHeap(filepath.Join(t.TempDir(), "missing-dir", "heap.prof")); err == nil {
		t.Fatal("expected error for unwritable path")
	}
}
