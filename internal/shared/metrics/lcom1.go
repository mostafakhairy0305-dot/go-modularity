package metrics

// MetricLCOM1 is the metric name for the LCOM1 cohesion metric.
const MetricLCOM1 = "lcom1"

// DefinitionLCOM1 versions the LCOM1 formula implemented by LCOM1.
const DefinitionLCOM1 = "go-modularity/lcom1-v1"

// LCOM1 computes Lack of Cohesion in Methods over unordered method pairs,
// where two methods "share" when their field-usage sets intersect:
//
//	LCOM1 = max(nonSharingPairs − sharingPairs, 0)
//
// Not applicable when methodCount < 2 or fieldCount == 0.
func LCOM1(nonSharingPairs, sharingPairs, methodCount, fieldCount int) MetricResult {
	if methodCount < 2 {
		return notApplicable(MetricLCOM1, ScopeType, DefinitionLCOM1,
			"type has fewer than 2 methods")
	}
	if fieldCount == 0 {
		return notApplicable(MetricLCOM1, ScopeType, DefinitionLCOM1, "type has no fields")
	}
	value := nonSharingPairs - sharingPairs
	if value < 0 {
		value = 0
	}
	return applicable(MetricLCOM1, ScopeType, DefinitionLCOM1, float64(value))
}
