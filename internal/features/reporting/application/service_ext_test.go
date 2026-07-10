package application_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	reporting "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

type bufSink struct{ buf *bytes.Buffer }

func (b bufSink) Open() (io.WriteCloser, error) { return nopCloser{b.buf}, nil }

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func report() gomodularity.Report {
	return gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{{
			Path:    "example.com/m/a",
			Metrics: []metrics.MetricResult{{Name: "abstractness", Scope: metrics.ScopePackage, Value: 0.5, Applicable: true, Definition: "d"}},
			Types: []gomodularity.TypeReport{{Name: "A", Metrics: []metrics.MetricResult{
				{Name: "amc", Scope: metrics.ScopeType, Value: 2, Applicable: true, Definition: "d"},
			}}},
		}},
	}
}

// Black-box: the text format includes the module and the type row.
func TestWriteText(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatText, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	out := sink.buf.String()
	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "A") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatJSON, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal(sink.buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got["schema_version"] != "1" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}
}

// Black-box: the CSV format starts with the canonical header and has a row per
// entity/metric.
func TestWriteCSV(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	if err := reporting.Write(report(), domain.FormatCSV, sink, domain.TextOptions{}); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(sink.buf).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("csv has %d rows, want header + data", len(records))
	}

	header := strings.Join(records[0], ",")
	if header != strings.Join(domain.CSVHeader(), ",") {
		t.Errorf("csv header = %q", header)
	}
}
