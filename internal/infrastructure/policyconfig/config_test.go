package policyconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/policyconfig"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// write drops content into a temp file and returns its path.
func write(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), policyconfig.FileName)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestLoadValid(t *testing.T) {
	t.Parallel()

	path := write(t, `
version: 1
package:
  types: 15
  efferent: 20
  metrics:
    distance: { min: 0.1, max: 0.6 }
type:
  fields: 12
  metrics:
    amc:  { max: 4 }
    camc: { min: 0.35 }
metrics:
  reusability: 0.6
`)

	policy, err := policyconfig.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if !policy.Package.Types.HasMax || policy.Package.Types.Max != 15 {
		t.Errorf("package.types = %+v", policy.Package.Types)
	}

	if !policy.Type.Fields.HasMax || policy.Type.Fields.Max != 12 {
		t.Errorf("type.fields = %+v", policy.Type.Fields)
	}

	// Top-level metrics remain the legacy/global form. A bare scalar is an upper bound.
	if l := policy.Metrics[metrics.MetricReusability]; !l.HasMax || l.HasMin || l.Max != 0.6 {
		t.Errorf("reusability scalar = %+v, want max 0.6 only", l)
	}

	if l := policy.TypeMetrics[metrics.MetricCAMC]; !l.HasMin || l.HasMax || l.Min != 0.35 {
		t.Errorf("camc = %+v, want min 0.35 only", l)
	}

	if l := policy.PackageMetrics[metrics.MetricDistance]; !l.HasMin || !l.HasMax {
		t.Errorf("distance = %+v, want both bounds", l)
	}
}

func TestLoadErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		content string
		want    string
	}{
		"unsupported version":    {"version: 2\n", "unsupported version"},
		"missing version":        {"package:\n  types: 5\n", "unsupported version"},
		"unknown structural key": {"version: 1\npackage:\n  bogus: 5\n", "field bogus not found"},
		"unknown metric":         {"version: 1\nmetrics:\n  nope: { max: 1 }\n", "unknown metric"},
		"wrong package metric":   {"version: 1\npackage:\n  metrics:\n    amc: { max: 1 }\n", "unknown package metric"},
		"wrong type metric":      {"version: 1\ntype:\n  metrics:\n    distance: { max: 1 }\n", "unknown type metric"},
		"unknown limit field":    {"version: 1\nmetrics:\n  amc: { maks: 4 }\n", "unknown limit field"},
		"empty limit mapping":    {"version: 1\nmetrics:\n  amc: {}\n", "must set max and/or min"},
		"min over max":           {"version: 1\nmetrics:\n  amc: { min: 5, max: 2 }\n", "min 5 exceeds max 2"},
		"malformed yaml":         {":\n  - broken\n:::", ""},
		"empty file":             {"", "config is empty"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := policyconfig.Load(write(t, tc.content))
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.want)
			}

			if tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	if _, err := policyconfig.Load(filepath.Join(t.TempDir(), "absent.yml")); err == nil {
		t.Error("loading a missing file: want error, got nil")
	}
}

func TestDiscover(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if _, ok := policyconfig.Discover(dir); ok {
		t.Error("discovered a config in an empty dir")
	}

	if err := os.WriteFile(filepath.Join(dir, policyconfig.FileName), []byte("version: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	path, ok := policyconfig.Discover(dir)
	if !ok {
		t.Fatal("did not discover the config")
	}

	if filepath.Base(path) != policyconfig.FileName {
		t.Errorf("discovered path = %q", path)
	}
}
