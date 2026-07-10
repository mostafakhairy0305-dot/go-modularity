// Package profiling wraps runtime/pprof for the CLI's --cpu-profile and
// --memory-profile flags.
package profiling

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// StartCPU begins CPU profiling into path and returns a stop function that
// finishes the profile and closes the file.
func StartCPU(path string) (stop func() error, err error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("start cpu profile: %w", err)
	}
	return func() error {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			return fmt.Errorf("close cpu profile: %w", err)
		}
		return nil
	}, nil
}

// WriteHeap writes a heap profile to path after forcing a garbage collection
// so the profile reflects live allocations.
func WriteHeap(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create memory profile: %w", err)
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		_ = f.Close()
		return fmt.Errorf("write memory profile: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close memory profile: %w", err)
	}
	return nil
}
