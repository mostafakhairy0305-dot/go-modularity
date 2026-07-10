package application

import (
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// White-box: the request→fact-options mapping.
func TestCollectOptionsMapping(t *testing.T) {
	t.Parallel()

	fo := collectOptions(inbound.Options{
		Directory: "d", Patterns: []string{"./..."}, IncludeTests: true,
		IncludeGenerated: true, BuildTags: []string{"tag"}, Workers: 3, ContinueOnError: true,
	})
	if fo.Directory != "d" || !fo.IncludeTests || !fo.IncludeGenerated ||
		fo.Workers != 3 || !fo.ContinueOnError || len(fo.BuildTags) != 1 {
		t.Fatalf("collectOptions = %+v", fo)
	}
}

// White-box: the reusability service is built only when needed.
func TestNewReusabilityServiceGating(t *testing.T) {
	t.Parallel()

	svc, err := newReusabilityService(map[string]bool{}, metrics.DefaultReusabilityWeights())
	if err != nil || svc != nil {
		t.Fatalf("no reusability/cbo selected → nil service; got %v err %v", svc, err)
	}

	svc, err = newReusabilityService(map[string]bool{metrics.MetricReusability: true}, metrics.DefaultReusabilityWeights())
	if err != nil || svc == nil {
		t.Fatalf("reusability selected → service; got %v err %v", svc, err)
	}

	if _, err := newReusabilityService(map[string]bool{metrics.MetricCBO: true},
		metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1}); err == nil {
		t.Fatal("bad weights should fail")
	}
}

// White-box: package metrics are skipped when none is displayed.
func TestComputeArchitectureGating(t *testing.T) {
	t.Parallel()

	if got := computeArchitecture(nil, map[string]bool{}, inbound.Options{}); got != nil {
		t.Fatalf("no package metric selected → nil; got %v", got)
	}
}

// White-box: nameSet builds a membership set.
func TestNameSet(t *testing.T) {
	t.Parallel()

	s := nameSet([]string{"a", "b", "a"})
	if len(s) != 2 || !s["a"] || !s["b"] {
		t.Fatalf("nameSet = %v", s)
	}
}
