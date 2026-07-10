package gomodularity

import (
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

func TestWithDefaults(t *testing.T) {
	cfg := Config{}.withDefaults()
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "./..." {
		t.Fatalf("patterns = %v", cfg.Patterns)
	}
	if cfg.DependencyScope != DependencyScopeModule {
		t.Fatalf("scope = %q", cfg.DependencyScope)
	}
	if cfg.FieldUsageMode != FieldUsageDirect {
		t.Fatalf("field usage = %q", cfg.FieldUsageMode)
	}
	if cfg.ReusabilityWeights != metrics.DefaultReusabilityWeights() {
		t.Fatalf("weights = %+v", cfg.ReusabilityWeights)
	}
	// Default display set excludes CBO.
	for _, m := range cfg.SelectedMetrics {
		if m == MetricCBO {
			t.Fatal("cbo must not be in the default display set")
		}
	}
	if len(cfg.SelectedMetrics) != len(AllMetrics())-1 {
		t.Fatalf("selected = %v", cfg.SelectedMetrics)
	}
}

func TestValidate(t *testing.T) {
	valid := Config{}.withDefaults()
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}

	bad := valid
	bad.DependencyScope = "galaxy"
	if err := bad.validate(); err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = valid
	bad.FieldUsageMode = "psychic"
	if err := bad.validate(); err == nil {
		t.Fatal("invalid field usage accepted")
	}

	bad = valid
	bad.SelectedMetrics = []MetricName{"made-up"}
	if err := bad.validate(); err == nil {
		t.Fatal("unknown metric accepted")
	}

	bad = valid
	bad.Patterns = []string{""}
	if err := bad.validate(); err == nil {
		t.Fatal("empty pattern accepted")
	}

	bad = valid
	bad.ReusabilityWeights = ReusabilityWeights{Cohesion: -0.5, Coupling: 1}
	if err := bad.validate(); err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestMetricClosure(t *testing.T) {
	closure := MetricClosure([]MetricName{MetricReusability, MetricDistance})
	want := map[MetricName]bool{
		MetricAMC: true, MetricLCOM96b: true, MetricCBO: true, MetricReusability: true,
		MetricAbstractness: true, MetricInstability: true, MetricDistance: true,
	}
	if len(closure) != len(want) {
		t.Fatalf("closure = %v", closure)
	}
	for _, m := range closure {
		if !want[m] {
			t.Fatalf("unexpected metric %q in closure %v", m, closure)
		}
	}
}
