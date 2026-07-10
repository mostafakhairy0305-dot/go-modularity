package metrics

// MetricLCOM96b is the metric name for the LCOM96b cohesion metric.
const MetricLCOM96b = "lcom96b"

// DefinitionLCOM96b versions the LCOM96b formula implemented by LCOM96b.
//
// This is the method-field matrix-density variant, chosen because it is
// defined at methodCount == 1 (unlike Henderson-Sellers LCOM*).
const DefinitionLCOM96b = "go-modularity/lcom96b-v1"

// LCOM96b computes the matrix-density lack-of-cohesion metric:
//
//	LCOM96b = 1 − totalMethodFieldAccesses / (fieldCount × methodCount)
//
// totalMethodFieldAccesses is the number of 1-cells in the method-field
// matrix (each method contributes each distinct field it uses once). The
// result is in [0, 1]. Not applicable when fieldCount == 0 or
// methodCount == 0.
func LCOM96b(totalMethodFieldAccesses, fieldCount, methodCount int) MetricResult {
	if fieldCount == 0 {
		return notApplicable(MetricLCOM96b, ScopeType, DefinitionLCOM96b, "type has no fields")
	}
	if methodCount == 0 {
		return notApplicable(MetricLCOM96b, ScopeType, DefinitionLCOM96b, "type has no methods")
	}
	density := float64(totalMethodFieldAccesses) / (float64(fieldCount) * float64(methodCount))
	return applicable(MetricLCOM96b, ScopeType, DefinitionLCOM96b, 1-density)
}
