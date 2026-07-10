package sinks

import (
	"bufio"
	"io"
	"os"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/ports/outbound"
)

// Compile-time proof that both adapters satisfy the reporting outbound port.
var (
	_ outbound.Sink = StdoutSink{}
	_ outbound.Sink = FileSink{}
)

// StdoutSink writes the report to standard output. Closing the returned
// stream flushes it — stdout itself stays open, the process owns it.
type StdoutSink struct{}

// Open returns a buffered stdout stream.
func (StdoutSink) Open() (io.WriteCloser, error) {
	return stdoutStream{bufio.NewWriter(os.Stdout)}, nil
}

// stdoutStream buffers report output to standard output; Write is the
// embedded buffered writer's.
type stdoutStream struct{ *bufio.Writer }

// Close flushes buffered output without closing stdout.
func (s stdoutStream) Close() error { return s.Flush() }

// FileSink writes the report to a file, truncating any existing content.
type FileSink struct {
	Path string
}

// Open creates the destination file.
func (s FileSink) Open() (io.WriteCloser, error) {
	return os.Create(s.Path)
}
