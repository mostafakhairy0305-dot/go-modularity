package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

func TestNewRejectsInvalidSettings(t *testing.T) {
	t.Parallel()

	_, err := New(Settings{DependencyScope: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid dependency-scope")
	}

	_, err = New(Settings{FieldUsage: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid field-usage")
	}
}

func TestNewAcceptsDefaults(t *testing.T) {
	t.Parallel()

	a, err := New(Settings{})
	if err != nil {
		t.Fatal(err)
	}

	if a.Name != Name {
		t.Fatalf("Name = %q, want %q", a.Name, Name)
	}
}

func TestRunnerLoadGroupsViolations(t *testing.T) {
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixture")
	policyPath := filepath.Join(t.TempDir(), "policy.yml")

	if err := os.WriteFile(policyPath, []byte("---\nversion: 1\ntype:\n  methods:\n    max: 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := newRunner(Settings{
		Directory: fixtureDir,
		Config:    policyPath,
		Patterns:  []string{"./isolated"},
	}.withDefaults())

	r.load()
	if r.err != nil {
		t.Fatal(r.err)
	}

	got := r.byPkg["example.com/fixture/isolated"]
	if len(got) == 0 {
		t.Fatal("expected violations for isolated with methods max: 0")
	}

	foundMethods := false
	for _, v := range got {
		if v.Key == policydomain.KeyMethods && v.Type == "Value" {
			foundMethods = true
		}
	}

	if !foundMethods {
		t.Fatalf("expected methods violation on Value, got %#v", got)
	}
}

func TestFormatViolation(t *testing.T) {
	t.Parallel()

	msg := formatViolation(policydomain.Violation{
		Package:    "example.com/p",
		Type:       "T",
		Key:        "methods",
		Value:      3,
		Comparator: policydomain.ComparatorMax,
		Threshold:  0,
	})

	want := "example.com/p.T (type): methods 3 exceeds max 0"
	if msg != want {
		t.Fatalf("formatViolation = %q, want %q", msg, want)
	}
}

func TestTypePosAndPackagePos(t *testing.T) {
	t.Parallel()

	src := "package p\n\ntype Widget struct{}\n"
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	pass := &analysis.Pass{Files: []*ast.File{file}, Fset: fset}

	if pos := typePos(pass, "Widget"); pos == token.NoPos {
		t.Fatal("typePos(Widget) = NoPos")
	}

	if pos := typePos(pass, "Missing"); pos != file.Package {
		t.Fatalf("typePos(Missing) = %v, want package clause", pos)
	}

	if pos := packagePos(pass); pos != file.Package {
		t.Fatalf("packagePos = %v, want %v", pos, file.Package)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
