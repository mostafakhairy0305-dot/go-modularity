package application

import (
	"testing"

	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

func fieldSet(size int, idx ...int) bitset.FieldSet {
	s := bitset.NewFieldSet(size)
	for _, i := range idx {
		s.Set(i)
	}
	return s
}

// White-box: two methods over disjoint fields are maximally incohesive.
func TestComputeForTypeDisjoint(t *testing.T) {
	t.Parallel()
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0)},
			{Name: "M2", FieldsUsed: fieldSet(2, 1)},
		},
	}
	got := ComputeForType(tf, false)

	// One non-sharing pair, zero sharing → LCOM1 = max(1-0, 0) = 1.
	if !got.LCOM1.Applicable || got.LCOM1.Value != 1 {
		t.Errorf("LCOM1 = %+v, want 1", got.LCOM1)
	}
	// No connected pairs → TCC = 0.
	if !got.TCC.Applicable || got.TCC.Value != 0 {
		t.Errorf("TCC = %+v, want 0", got.TCC)
	}
}

// White-box: transitive mode propagates field usage through sibling calls,
// making the two methods share and become cohesive.
func TestComputeForTypeTransitive(t *testing.T) {
	t.Parallel()
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fieldSet(2, 0), CalledSiblings: []int{1}},
			{Name: "M2", FieldsUsed: fieldSet(2, 1)},
		},
	}
	direct := ComputeForType(tf, false)
	transitive := ComputeForType(tf, true)

	if direct.TCC.Value != 0 {
		t.Errorf("direct TCC = %v, want 0", direct.TCC.Value)
	}
	if transitive.TCC.Value != 1 {
		t.Errorf("transitive TCC = %v, want 1 (M1 now shares via M2)", transitive.TCC.Value)
	}
}
