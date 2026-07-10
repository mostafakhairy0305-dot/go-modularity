package goloader_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/goloader"
)

func fixtureDir() string { return filepath.Join("..", "..", "..", "testdata", "fixture") }

// Black-box: the loader extracts real facts from the fixture module through
// the outbound port.
func TestLoadFixture(t *testing.T) {
	t.Parallel()
	mod, pkgs, err := goloader.New().Load(context.Background(), outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./..."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if mod != "example.com/fixture" {
		t.Fatalf("module = %q", mod)
	}

	var orders *domain.PackageExtract
	for i := range pkgs {
		if pkgs[i].Path == "example.com/fixture/orders" {
			orders = &pkgs[i]
		}
	}
	if orders == nil {
		t.Fatal("orders package not extracted")
	}
	if !orders.InModule {
		t.Error("orders should be in-module")
	}

	var order *domain.TypeExtract
	for i := range orders.Types {
		if orders.Types[i].Name == "Order" {
			order = &orders.Types[i]
		}
	}
	if order == nil {
		t.Fatal("Order type not extracted")
	}
	if len(order.Methods) != 3 {
		t.Errorf("Order methods = %d, want 3", len(order.Methods))
	}
	if !slices.Contains(order.ReferencedTypeKeys, "example.com/fixture/store.Store") {
		t.Errorf("Order refs = %v, want to include store.Store", order.ReferencedTypeKeys)
	}
}

// Black-box: a pattern matching nothing is an error.
func TestLoadNoMatch(t *testing.T) {
	t.Parallel()
	_, _, err := goloader.New().Load(context.Background(), outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./does-not-exist"},
	})
	if err == nil {
		t.Fatal("expected error for a pattern matching no packages")
	}
}
