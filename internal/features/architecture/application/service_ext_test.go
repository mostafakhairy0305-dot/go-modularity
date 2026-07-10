package application_test

import (
	"testing"

	architecture "github.com/mostafakhairy0305-dot/go-modularity/internal/features/architecture/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/architecture/domain"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
)

// Black-box: an isolated concrete package (no in-scope deps) sits in the Zone
// of Pain — instability defaults to 0, abstractness 0, so distance is 1.
func TestComputeForPackagesIsolated(t *testing.T) {
	t.Parallel()
	facts := &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{ID: 0, Path: "example.com/m/solo", InModule: true, TypeIDs: []int{0}},
		},
		Types: []typefacts.TypeFacts{
			{ID: 0, PackageID: 0, Name: "Solo", Kind: typefacts.KindStruct},
		},
	}

	got := architecture.ComputeForPackages(facts, domain.Scope("project"))
	if len(got) != 1 {
		t.Fatalf("got %d results", len(got))
	}
	if !got[0].Distance.Applicable || got[0].Distance.Value != 1 {
		t.Fatalf("isolated distance = %+v, want 1", got[0].Distance)
	}
	if !got[0].Instability.Applicable || got[0].Instability.Value != 0 {
		t.Fatalf("isolated instability = %+v, want 0", got[0].Instability)
	}
}
