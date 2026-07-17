package analyzer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity/gomodularity"
	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
)

// PackageSettings configures limits evaluated once per package.
type PackageSettings struct {
	Types           *LimitSettings           `json:"types"`
	ExportedFuncs   *LimitSettings           `json:"exported_funcs"`
	UnexportedFuncs *LimitSettings           `json:"unexported_funcs"`
	Afferent        *LimitSettings           `json:"afferent"`
	Efferent        *LimitSettings           `json:"efferent"`
	Metrics         map[string]LimitSettings `json:"metrics"`
}

// TypeSettings configures limits evaluated once per named type.
type TypeSettings struct {
	Fields  *LimitSettings           `json:"fields"`
	Methods *LimitSettings           `json:"methods"`
	Metrics map[string]LimitSettings `json:"metrics"`
}

// LimitSettings is an optional lower bound, upper bound, or both. When decoded
// from golangci-lint settings, a bare number is shorthand for a maximum.
type LimitSettings struct {
	Max *float64 `json:"max"`
	Min *float64 `json:"min"`
}

// UnmarshalJSON accepts either a bare numeric maximum or a strict
// {"min": ..., "max": ...} object.
func (l *LimitSettings) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return errors.New("limit must be a number or a {min, max} mapping")
	}

	var scalar float64
	if err := json.Unmarshal(trimmed, &scalar); err == nil {
		l.Max = &scalar
		l.Min = nil

		return nil
	}

	var bounds struct {
		Max *float64 `json:"max"`
		Min *float64 `json:"min"`
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bounds); err != nil {
		return fmt.Errorf("limit must be a number or a {min, max} mapping: %w", err)
	}

	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}

	if bounds.Max == nil && bounds.Min == nil {
		return errors.New("limit must set max and/or min")
	}

	l.Max = bounds.Max
	l.Min = bounds.Min

	return nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("limit must contain exactly one JSON value")
		}

		return fmt.Errorf("decoding limit: %w", err)
	}

	return nil
}

// policy returns the inline policy. With no policy keys configured, the
// recommended defaults apply. It never reads or discovers .modularity.yml.
func (s Settings) policy() (policydomain.Policy, error) {
	if s.Package == nil && s.Type == nil && s.Metrics == nil {
		return policydomain.DefaultPolicy(), nil
	}

	policy := policydomain.Policy{
		Metrics:        make(map[string]policydomain.Limit, len(s.Metrics)),
		PackageMetrics: make(map[string]policydomain.Limit),
		TypeMetrics:    make(map[string]policydomain.Limit),
	}

	if err := applyPackageSettings(&policy, s.Package); err != nil {
		return policydomain.Policy{}, err
	}

	if err := applyTypeSettings(&policy, s.Type); err != nil {
		return policydomain.Policy{}, err
	}

	if err := copyLimitSettings(policy.Metrics, s.Metrics); err != nil {
		return policydomain.Policy{}, err
	}

	if err := policydomain.Validate(policy); err != nil {
		return policydomain.Policy{}, err
	}

	return policy, nil
}

func applyPackageSettings(policy *policydomain.Policy, settings *PackageSettings) error {
	if settings == nil {
		return nil
	}

	policy.Package.Types = settings.Types.toLimit()
	policy.Package.ExportedFuncs = settings.ExportedFuncs.toLimit()
	policy.Package.UnexportedFuncs = settings.UnexportedFuncs.toLimit()
	policy.Package.Afferent = settings.Afferent.toLimit()
	policy.Package.Efferent = settings.Efferent.toLimit()
	policy.PackageMetrics = make(map[string]policydomain.Limit, len(settings.Metrics))

	return copyLimitSettings(policy.PackageMetrics, settings.Metrics)
}

func applyTypeSettings(policy *policydomain.Policy, settings *TypeSettings) error {
	if settings == nil {
		return nil
	}

	policy.Type.Fields = settings.Fields.toLimit()
	policy.Type.Methods = settings.Methods.toLimit()
	policy.TypeMetrics = make(map[string]policydomain.Limit, len(settings.Metrics))

	return copyLimitSettings(policy.TypeMetrics, settings.Metrics)
}

func (l *LimitSettings) toLimit() policydomain.Limit {
	if l == nil {
		return policydomain.Limit{}
	}

	limit := policydomain.Limit{}
	if l.Max != nil {
		limit.Max = *l.Max
		limit.HasMax = true
	}
	if l.Min != nil {
		limit.Min = *l.Min
		limit.HasMin = true
	}

	return limit
}

func copyLimitSettings(
	destination map[string]policydomain.Limit,
	source map[string]LimitSettings,
) error {
	for name, settings := range source {
		limit := settings.toLimit()
		if !limit.HasMax && !limit.HasMin {
			return fmt.Errorf("%s: limit must set max and/or min", name)
		}

		destination[name] = limit
	}

	return nil
}

// gatedMetrics unions the policy's constrained metrics into the display set so
// every gated metric is computed — a metric absent from the report cannot be
// checked.
func gatedMetrics(policy policydomain.Policy) []gomodularity.MetricName {
	base := gomodularity.DefaultMetrics()
	present := make(map[gomodularity.MetricName]bool, len(base))
	out := append([]gomodularity.MetricName(nil), base...)

	for _, name := range base {
		present[name] = true
	}

	for _, name := range policydomain.MetricNames(policy) {
		metric := gomodularity.MetricName(name)
		if present[metric] {
			continue
		}

		out = append(out, metric)
		present[metric] = true
	}

	return out
}
