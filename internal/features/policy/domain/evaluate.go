package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
)

// Comparator identifies which bound a violation crossed.
type Comparator string

const (
	// ComparatorMax marks a value that exceeded an upper bound.
	ComparatorMax Comparator = "max"
	// ComparatorMin marks a value that fell below a lower bound.
	ComparatorMin Comparator = "min"
)

// Violation is one broken condition: a field's actual value against the bound
// it crossed. Type is empty for package-scope conditions.
type Violation struct {
	Package    string     // package import path
	Type       string     // type name; empty for package-scope conditions
	Key        string     // condition key: structural key or metric name
	Value      float64    // the entity's actual value
	Comparator Comparator // which bound was crossed
	Threshold  float64    // the bound's value
}

// Evaluate checks a report against a policy and returns the violations. The
// result is deterministic: packages are already sorted by path and types by
// name, structural conditions precede metric conditions, and metrics keep the
// report's fixed order. A metric condition is skipped for any entity where the
// metric is not applicable, so n/a cells never produce false positives.
func Evaluate(report gomodularity.Report, policy Policy) []Violation {
	var violations []Violation

	for i := range report.Packages {
		pkg := &report.Packages[i]

		check(&violations, pkg.Path, "", KeyTypes, float64(len(pkg.Types)), policy.Package.Types)
		check(&violations, pkg.Path, "", KeyExportedFuncs, float64(pkg.ExportedFuncs), policy.Package.ExportedFuncs)
		check(&violations, pkg.Path, "", KeyUnexportedFuncs, float64(pkg.UnexportedFuncs), policy.Package.UnexportedFuncs)
		check(&violations, pkg.Path, "", KeyAfferent, float64(pkg.Afferent), policy.Package.Afferent)
		check(&violations, pkg.Path, "", KeyEfferent, float64(pkg.Efferent), policy.Package.Efferent)

		for _, result := range pkg.Metrics {
			if !result.Applicable {
				continue
			}

			if limit, ok := packageMetricLimit(policy, result.Name); ok {
				check(&violations, pkg.Path, "", result.Name, result.Value, limit)
			}
		}

		for j := range pkg.Types {
			typ := &pkg.Types[j]

			check(&violations, pkg.Path, typ.Name, KeyFields, float64(typ.Fields), policy.Type.Fields)
			check(&violations, pkg.Path, typ.Name, KeyMethods, float64(typ.Methods), policy.Type.Methods)

			for _, result := range typ.Metrics {
				if !result.Applicable {
					continue
				}

				if limit, ok := typeMetricLimit(policy, result.Name); ok {
					check(&violations, pkg.Path, typ.Name, result.Name, result.Value, limit)
				}
			}
		}
	}

	return violations
}

// check appends a violation for each bound the value crosses.
func check(violations *[]Violation, pkg, typ, key string, value float64, limit Limit) {
	if limit.HasMax && value > limit.Max {
		*violations = append(*violations, Violation{
			Package: pkg, Type: typ, Key: key, Value: value,
			Comparator: ComparatorMax, Threshold: limit.Max,
		})
	}

	if limit.HasMin && value < limit.Min {
		*violations = append(*violations, Violation{
			Package: pkg, Type: typ, Key: key, Value: value,
			Comparator: ComparatorMin, Threshold: limit.Min,
		})
	}
}

// FormatViolations renders violations as a human-readable summary. The empty
// slice yields the empty string, so callers can print unconditionally.
func FormatViolations(violations []Violation) string {
	if len(violations) == 0 {
		return ""
	}

	var b strings.Builder

	noun := "violations"
	if len(violations) == 1 {
		noun = "violation"
	}

	fmt.Fprintf(&b, "policy: %d %s\n", len(violations), noun)

	for _, v := range violations {
		where := v.Package + " (package)"
		if v.Type != "" {
			where = v.Package + "." + v.Type + " (type)"
		}

		relation := "exceeds max"
		if v.Comparator == ComparatorMin {
			relation = "is below min"
		}

		fmt.Fprintf(&b, "  %s: %s %s %s %s\n",
			where, v.Key, formatNumber(v.Value), relation, formatNumber(v.Threshold))
	}

	return b.String()
}

// formatNumber prints integers without a fraction and other values with two
// decimals, matching the report's cell formatting.
func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
