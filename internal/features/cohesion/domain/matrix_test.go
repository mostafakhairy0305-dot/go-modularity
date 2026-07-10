package domain

import (
	"testing"

	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

func fieldSet(size int, indices ...int) bitset.FieldSet {
	s := bitset.NewFieldSet(size)
	for _, i := range indices {
		s.Set(i)
	}
	return s
}

func TestCountPairsSmallPath(t *testing.T) {
	sets := []bitset.FieldSet{
		fieldSet(3, 0),
		fieldSet(3, 0, 1),
		fieldSet(3, 2),
	}
	counts := CountPairs(sets, 3)
	if counts.Sharing != 1 || counts.NonSharing != 2 {
		t.Fatalf("counts = %+v, want sharing 1, non-sharing 2", counts)
	}
}

func TestCountPairsGeneralPath(t *testing.T) {
	// 70 fields forces the multi-word path.
	sets := []bitset.FieldSet{
		fieldSet(70, 69),
		fieldSet(70, 69, 1),
		fieldSet(70, 5),
	}
	counts := CountPairs(sets, 70)
	if counts.Sharing != 1 || counts.NonSharing != 2 {
		t.Fatalf("counts = %+v, want sharing 1, non-sharing 2", counts)
	}
}

func TestCountPairsDegenerate(t *testing.T) {
	if counts := CountPairs(nil, 0); counts.Sharing != 0 || counts.NonSharing != 0 {
		t.Fatalf("counts = %+v", counts)
	}
	// Methods with no field usage never share.
	counts := CountPairs([]bitset.FieldSet{{}, {}}, 0)
	if counts.Sharing != 0 || counts.NonSharing != 1 {
		t.Fatalf("counts = %+v, want non-sharing 1", counts)
	}
}

func TestEffectiveFieldSetsTransitive(t *testing.T) {
	// a uses field 0 and calls b; b calls c; c uses field 1.
	facts := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "x"}, {Name: "y"}},
		Methods: []typefacts.MethodFacts{
			{Name: "a", FieldsUsed: fieldSet(2, 0), CalledSiblings: []int{1}},
			{Name: "b", FieldsUsed: fieldSet(2), CalledSiblings: []int{2}},
			{Name: "c", FieldsUsed: fieldSet(2, 1)},
		},
	}

	direct := EffectiveFieldSets(facts, false)
	if bitset.Count(direct[0]) != 1 || bitset.Count(direct[1]) != 0 {
		t.Fatalf("direct sets changed: %v", direct)
	}

	transitive := EffectiveFieldSets(facts, true)
	if !bitset.Contains(transitive[0], 1) {
		t.Fatal("a should reach c's field through b (fixpoint)")
	}
	if bitset.Count(transitive[1]) != 1 || !bitset.Contains(transitive[1], 1) {
		t.Fatal("b should absorb c's field")
	}
	// The original facts stay untouched.
	if bitset.Contains(facts.Methods[0].FieldsUsed, 1) {
		t.Fatal("transitive mode mutated the extracted facts")
	}
}

func TestParamMatrix(t *testing.T) {
	methods := []typefacts.MethodFacts{
		{ParamTypeKeys: []string{"int", "string"}},
		{ParamTypeKeys: []string{"int"}},
		{},
	}
	oneCells, distinct := ParamMatrix(methods)
	if oneCells != 3 || distinct != 2 {
		t.Fatalf("oneCells = %d, distinct = %d; want 3, 2", oneCells, distinct)
	}
}

func TestTotalFieldAccesses(t *testing.T) {
	sets := []bitset.FieldSet{fieldSet(3, 0, 1), fieldSet(3, 1), {}}
	if got := TotalFieldAccesses(sets); got != 3 {
		t.Fatalf("total = %d, want 3", got)
	}
}
