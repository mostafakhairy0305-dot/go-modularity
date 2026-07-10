// Package outbound declares the reporting feature's outbound port: the
// destination a rendered report is written to.
package outbound

import "io"

// Sink provides the output stream for one rendered report.
type Sink interface {
	// Open returns the stream; the caller closes it when rendering is done.
	Open() (io.WriteCloser, error)
}
