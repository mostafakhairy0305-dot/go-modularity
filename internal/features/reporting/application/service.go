package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Write renders the report in the given format into the sink. Options are
// read only by the text format.
func Write(report gomodularity.Report, format domain.Format, sink outbound.Sink, opts domain.TextOptions) error {
	w, err := sink.Open()
	if err != nil {
		return err
	}

	renderErr := render(w, report, format, opts)
	if renderErr != nil {
		_ = w.Close()

		return renderErr
	}

	return w.Close()
}

func render(w io.Writer, report gomodularity.Report, format domain.Format, opts domain.TextOptions) error {
	switch format {
	case domain.FormatText:
		_, err := io.WriteString(w, domain.Text(report, opts))

		return err
	case domain.FormatJSON:
		return renderJSON(w, report)
	case domain.FormatCSV:
		cw := csv.NewWriter(w)
		err := cw.Write(domain.CSVHeader())
		if err != nil {
			return err
		}

		err = cw.WriteAll(domain.CSVRecords(report))
		if err != nil {
			return err
		}

		cw.Flush()

		return cw.Error()
	case domain.FormatWeb:
		return renderWeb(w, report)
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}

// jsonReport mirrors the versioned report schema (§ output). Metric maps
// are orderedMetrics so keys always appear in the fixed metric order.
type jsonReport struct {
	// SchemaVersion is the report schema version.
	SchemaVersion string `json:"schema_version"`
	// Tool identifies the producing tool.
	Tool jsonTool `json:"tool"`
	// Packages are the analyzed packages in report order.
	Packages []jsonPackage `json:"packages"`
}

// String summarizes the report envelope for debugging.
func (r jsonReport) String() string {
	return fmt.Sprintf("schema %s, tool %v, %d packages", r.SchemaVersion, r.Tool, len(r.Packages))
}

type jsonTool struct {
	// Name is the tool's canonical name.
	Name string `json:"name"`
	// Version is the tool's build version.
	Version string `json:"version"`
}

type jsonPackage struct {
	// Path is the package's import path.
	Path string `json:"path"`
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int `json:"afferent"`
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int `json:"efferent"`
	// Funcs counts the package's declared functions and methods.
	Funcs int `json:"funcs"`
	// Metrics maps metric names to results in the fixed order.
	Metrics orderedMetrics `json:"metrics"`
	// Types are the package's analyzed types in report order.
	Types []jsonType `json:"types"`
}

// String summarizes one package entry for debugging.
func (p jsonPackage) String() string {
	return fmt.Sprintf("%s: %d metrics, %d types", p.Path, len(p.Metrics), len(p.Types))
}

type jsonType struct {
	// Name is the type's declared name.
	Name string `json:"name"`
	// Fields is the struct field count (embedded fields count one).
	Fields int `json:"fields"`
	// Methods is the declared method count.
	Methods int `json:"methods"`
	// Metrics maps metric names to results in the fixed order.
	Metrics orderedMetrics `json:"metrics"`
}

// jsonMetric serializes one MetricResult. A non-applicable metric carries
// its reason and no value — never a fake zero.
type jsonMetric struct {
	// Scope is the kind of entity the metric describes.
	Scope string `json:"scope"`
	// Value is the metric value, present only when applicable.
	Value *float64 `json:"value,omitempty"`
	// Applicable reports whether the value may be read.
	Applicable bool `json:"applicable"`
	// Reason explains non-applicability or dropped components.
	Reason string `json:"reason,omitempty"`
	// Definition is the versioned formula identifier.
	Definition string `json:"definition"`
}

// orderedMetrics marshals as a JSON object keyed by metric name, preserving
// slice order (the fixed metric order).
type orderedMetrics []metrics.MetricResult

// MarshalJSON writes the object with keys in the fixed metric order.
func (m orderedMetrics) MarshalJSON() ([]byte, error) {
	return encodeOrderedMetrics(m)
}

// encodeOrderedMetrics assembles the ordered JSON object one name→metric
// pair at a time.
func encodeOrderedMetrics(results []metrics.MetricResult) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, r := range results {
		if i > 0 {
			buf.WriteByte(',')
		}

		err := encodeMetricEntry(&buf, r)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// encodeMetricEntry writes one name→metric pair. A non-applicable metric
// carries its reason and no value — never a fake zero.
func encodeMetricEntry(buf *bytes.Buffer, r metrics.MetricResult) error {
	key, err := json.Marshal(r.Name)
	if err != nil {
		return err
	}

	buf.Write(key)
	buf.WriteByte(':')

	out := jsonMetric{
		Scope:      string(r.Scope),
		Applicable: r.Applicable,
		Reason:     r.Reason,
		Definition: r.Definition,
	}
	if r.Applicable {
		value := r.Value
		out.Value = &value
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return err
	}

	buf.Write(encoded)

	return nil
}

// buildJSONReport maps the report onto the versioned JSON schema. It is
// shared by the JSON format and the web report's embedded payload.
func buildJSONReport(report gomodularity.Report) jsonReport {
	out := jsonReport{
		SchemaVersion: report.SchemaVersion,
		Tool:          jsonTool{Name: report.Tool.Name, Version: report.Tool.Version},
		Packages:      make([]jsonPackage, len(report.Packages)),
	}
	for i, pkg := range report.Packages {
		jp := jsonPackage{
			Path:     pkg.Path,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
			Funcs:    pkg.ExportedFuncs + pkg.UnexportedFuncs,
			Metrics:  orderedMetrics(pkg.Metrics),
			Types:    make([]jsonType, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			jp.Types[j] = jsonType{
				Name:    t.Name,
				Fields:  t.Fields,
				Methods: t.Methods,
				Metrics: orderedMetrics(t.Metrics),
			}
		}

		out.Packages[i] = jp
	}

	return out
}

func renderJSON(w io.Writer, report gomodularity.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(buildJSONReport(report))
}
