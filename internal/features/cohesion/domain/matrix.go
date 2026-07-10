package domain

import (
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

// EffectiveFieldSets returns each method's field-usage set. In direct mode
// this is the extracted usage as-is; in transitive mode usage is propagated
// through calls to sibling methods until a fixpoint.
func EffectiveFieldSets(t *typefacts.TypeFacts, transitive bool) []bitset.FieldSet {
	sets := make([]bitset.FieldSet, len(t.Methods))
	for i := range t.Methods {
		sets[i] = t.Methods[i].FieldsUsed
	}

	if !transitive || len(t.Fields) == 0 {
		return sets
	}

	for i := range sets {
		sets[i] = bitset.Clone(sets[i])
	}

	for changed := true; changed; {
		changed = false

		for i := range t.Methods {
			for _, j := range t.Methods[i].CalledSiblings {
				before := bitset.Count(sets[i])
				bitset.Union(sets[i], sets[j])

				if bitset.Count(sets[i]) != before {
					changed = true
				}
			}
		}
	}

	return sets
}

// PairCounts summarizes the unordered method pairs of a type. Two methods
// share when their field sets intersect; a pair is connected under exactly
// the same predicate, so Sharing doubles as TCC's connected-pair count.
type PairCounts struct {
	// Sharing counts pairs whose field-usage sets intersect.
	Sharing int
	// NonSharing counts pairs with disjoint field-usage sets.
	NonSharing int
}

// CountPairs counts sharing and non-sharing unordered method pairs. The
// O(k²) loop works on bitsets only — the single-word fast path applies
// whenever the type has at most 64 fields.
func CountPairs(sets []bitset.FieldSet, fieldCount int) PairCounts {
	k := len(sets)

	var counts PairCounts
	if k < 2 {
		return counts
	}

	if fieldCount <= 64 {
		small := make([]bitset.SmallFieldSet, k)
		for i, s := range sets {
			small[i] = bitset.Small(s)
		}

		for i := range k {
			for j := i + 1; j < k; j++ {
				if small[i].Intersects(small[j]) {
					counts.Sharing++
				} else {
					counts.NonSharing++
				}
			}
		}

		return counts
	}

	for i := range k {
		for j := i + 1; j < k; j++ {
			if bitset.Intersects(sets[i], sets[j]) {
				counts.Sharing++
			} else {
				counts.NonSharing++
			}
		}
	}

	return counts
}

// TotalFieldAccesses is the number of 1-cells in the method-field matrix:
// each method contributes each distinct field it uses once.
func TotalFieldAccesses(sets []bitset.FieldSet) int {
	total := 0
	for _, s := range sets {
		total += bitset.Count(s)
	}

	return total
}

// ParamMatrix summarizes the method × parameter-type occurrence matrix:
// oneCells is the number of 1-cells (each method counts each of its distinct
// parameter types once) and distinct the matrix width. This feeds CAMC and
// is independent of the TCC computation.
func ParamMatrix(methods []typefacts.MethodFacts) (oneCells, distinct int) {
	seen := make(map[string]struct{})

	for i := range methods {
		oneCells += len(methods[i].ParamTypeKeys)
		for _, key := range methods[i].ParamTypeKeys {
			seen[key] = struct{}{}
		}
	}

	return oneCells, len(seen)
}
