package application_test

import (
	"testing"

	cohesion "github.com/mostafakhairy0305-dot/go-modularity/internal/features/cohesion/application"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

// Black-box: two methods sharing a field form one connected pair, giving
// full TCC and zero LCOM1.
func TestComputeForTypeCohesive(t *testing.T) {
	t.Parallel()
	shared := bitset.NewFieldSet(1)
	shared.Set(0)
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}},
		Methods: []typefacts.MethodFacts{
			{Name: "Get", FieldsUsed: cloneSet(shared)},
			{Name: "Set", FieldsUsed: cloneSet(shared)},
		},
	}
	got := cohesion.ComputeForType(tf, false)
	if !got.TCC.Applicable || got.TCC.Value != 1 {
		t.Errorf("TCC = %+v, want 1", got.TCC)
	}
	if !got.LCOM1.Applicable || got.LCOM1.Value != 0 {
		t.Errorf("LCOM1 = %+v, want 0", got.LCOM1)
	}
}

func cloneSet(s bitset.FieldSet) bitset.FieldSet {
	return bitset.Clone(s)
}
