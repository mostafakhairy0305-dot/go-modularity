package analyzer

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
	"golang.org/x/tools/go/analysis"
)

func TestRunnerRunReportsPackageAndTypeViolations(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{byPkg: map[string][]policydomain.Violation{
		"example.com/p": {
			{Package: "example.com/p", Key: policydomain.KeyTypes, Value: 2, Comparator: policydomain.ComparatorMax, Threshold: 1},
			{Package: "example.com/p", Type: "Widget", Key: policydomain.KeyMethods, Value: 0.5, Comparator: policydomain.ComparatorMin, Threshold: 1.25},
		},
	}}
	r.once.Do(func() {})

	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Fset:   fset,
		Files:  []*ast.File{file},
		Pkg:    types.NewPackage("example.com/p", "p"),
		Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
	}

	if _, err := r.run(pass); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v, want two", diagnostics)
	}
	if diagnostics[0].Pos != file.Package {
		t.Errorf("package diagnostic position = %v, want %v", diagnostics[0].Pos, file.Package)
	}
	if !strings.Contains(diagnostics[1].Message, "is below min 1.25") {
		t.Errorf("type diagnostic = %q", diagnostics[1].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{err: sentinel}
	failing.once.Do(func() {})
	if _, err := failing.run(pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	r := newRunner(Settings{Config: filepath.Join(t.TempDir(), "missing.yml")}.withDefaults())
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "gomodularity policy") {
		t.Fatalf("policy load error = %v", r.err)
	}

	r = newRunner(Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Patterns:  []string{"./..."},
	}.withDefaults())
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "gomodularity analyze") {
		t.Fatalf("analysis load error = %v", r.err)
	}
}

func TestResolvePolicyDiscoveryAndDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	defaults, err := resolvePolicy("", "")
	if err != nil {
		t.Fatal(err)
	}
	if !defaults.Package.Types.HasMax {
		t.Fatal("default policy was not returned")
	}

	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(config, []byte("version: 1\npackage:\n  types: 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	discovered, err := resolvePolicy("", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !discovered.Package.Types.HasMax || discovered.Package.Types.Max != 3 {
		t.Fatalf("discovered types limit = %+v", discovered.Package.Types)
	}
}

func TestGatedMetricsAndEmptyPackagePosition(t *testing.T) {
	policy := policydomain.Policy{TypeMetrics: map[string]policydomain.Limit{
		metrics.MetricAMC: {Max: 3, HasMax: true},
		metrics.MetricCBO: {Max: 5, HasMax: true},
	}}
	got := gatedMetrics(policy)
	if !slices.Contains(got, gomodularity.MetricCBO) {
		t.Fatalf("gatedMetrics() = %v, want cbo", got)
	}

	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}
