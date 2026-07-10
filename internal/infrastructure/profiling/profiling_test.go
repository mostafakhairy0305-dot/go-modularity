package profiling

import (
	"os"
	"path/filepath"
	"testing"
)

// White-box: StartCPU writes a profile and the stop function closes it.
// (Not parallel — the CPU profiler is process-global.)
func TestStartCPUWritesProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.prof")

	stop, err := StartCPU(path)
	if err != nil {
		t.Fatal(err)
	}

	sum := 0
	for i := range 100000 {
		sum += i
	}

	_ = sum

	if err := stop(); err != nil {
		t.Fatal(err)
	}

	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("cpu profile not written: err=%v", err)
	}
}

// White-box: StartCPU surfaces file-creation errors.
func TestStartCPUBadPath(t *testing.T) {
	_, err := StartCPU(filepath.Join(t.TempDir(), "missing-dir", "cpu.prof"))
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}
